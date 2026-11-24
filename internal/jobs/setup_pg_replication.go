package jobs

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/google/uuid"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service/secrets"
	"go.vervstack.ru/Velez/internal/service/user_errors"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/utils/common"
)

const (
	pgReplicationDriver         = "postgres"
	SetupPgReplicationAction    = "setup_pg_replication"
	PgReplicationModePhysical   = "physical"
	pgStreamingTimeout          = 60 * time.Second
	pgStreamingPollInterval     = time.Second
	pgReplicaSlotMaxLength      = 63
	pgReplicaSlotHashBytes      = 8
	pgReplicationMajorVersion   = 18
	pgServerVersionMajorDivisor = 10000
	pgTargetEmptyQuery          = `
SELECT pg_is_in_recovery(), EXISTS (
    SELECT 1
    FROM pg_class c
    JOIN pg_namespace n ON n.oid = c.relnamespace
    WHERE c.relkind IN ('r', 'p', 'f', 'm')
      AND n.nspname NOT IN ('pg_catalog', 'information_schema')
      AND n.nspname !~ '^pg_(toast|temp_)'
)`
)

type setupPgReplicationHandler struct {
	dataStorage storage.Storage
	secrets     secrets.Store
	docker      node_clients.Docker
}

func NewSetupPgReplicationHandler(
	dataStorage storage.Storage, secretsStore secrets.Store, docker node_clients.Docker,
) TaskHandler {
	return &setupPgReplicationHandler{dataStorage: dataStorage, secrets: secretsStore, docker: docker}
}

func (h *setupPgReplicationHandler) Action() string { return SetupPgReplicationAction }

func (h *setupPgReplicationHandler) NewContext() TaskContext {
	return &velez_api.SetupPgReplicationTaskPayload{}
}

func (h *setupPgReplicationHandler) BuildJobs(taskCtx TaskContext) []NamedJob {
	payload, ok := taskCtx.(*velez_api.SetupPgReplicationTaskPayload)
	if !ok {
		panic("setup_pg_replication: BuildJobs called with mismatched TaskContext type")
	}

	setup := &pgReplicationSetup{
		dataStorage: h.dataStorage,
		secrets:     h.secrets,
		docker:      h.docker,
		payload:     payload,
	}

	return []NamedJob{
		{Name: "resolve_endpoints", Job: pgReplicationStep(setup.resolveEndpoints)},
		{Name: "check_source_ready", Job: pgReplicationStep(setup.checkSourceReady)},
		{Name: "check_target_empty", Job: pgReplicationStep(setup.checkTargetEmpty)},
		{Name: "create_slot", Job: &pgReplicationPrepareSourceJob{setup: setup}},
		{Name: "stop_target", Job: &pgReplicationStopTargetJob{setup: setup}},
		{Name: "basebackup", Job: pgReplicationStep(setup.basebackup)},
		{Name: "start_target", Job: pgReplicationStep(setup.startTarget)},
		{Name: "verify_streaming", Job: pgReplicationStep(setup.verifyStreaming)},
	}
}

func FindPgInstanceContainerId(ctx context.Context, docker node_clients.Docker, name string) (string, error) {
	req := &velez_api.ListSmerds_Request{
		Label: map[string]string{
			labels.VervServiceLabel:   name,
			labels.PgaasInstanceLabel: labelTrueValue,
		},
	}

	containers, err := docker.ListContainers(ctx, req, "")
	if err != nil {
		return "", rerrors.Wrap(err, "error finding pg instance container")
	}

	switch len(containers) {
	case 0:
		return "", rerrors.Wrap(user_errors.ErrContainerNotFoundInEnvironment)
	case 1:
		return containers[0].ID, nil
	default:
		return "", rerrors.Wrap(user_errors.ErrPgReplicationContainerAmbiguous)
	}
}

type pgReplicationSetup struct {
	dataStorage storage.Storage
	secrets     secrets.Store
	docker      node_clients.Docker
	payload     *velez_api.SetupPgReplicationTaskPayload
}

func (s *pgReplicationSetup) resolveEndpoints(ctx context.Context) error {
	if s.payload.GetMode() != PgReplicationModePhysical {
		return rerrors.Wrap(user_errors.ErrPgReplicationMode)
	}

	if s.payload.GetSourceInstance() == s.payload.GetTargetInstance() {
		return rerrors.Wrap(user_errors.ErrPgReplicationSameInstance)
	}

	err := uuid.Validate(s.payload.GetOperationId())
	if err != nil || s.payload.GetTargetContainerId() == "" {
		return rerrors.Wrap(user_errors.ErrPgReplicationOperationMissing)
	}

	_, err = s.resolve(ctx, s.payload.GetSourceInstance())
	if err != nil {
		return rerrors.Wrap(err, "error resolving source instance")
	}

	_, err = s.resolve(ctx, s.payload.GetTargetInstance())
	if err != nil {
		return rerrors.Wrap(err, "error resolving target instance")
	}

	return nil
}

func (s *pgReplicationSetup) checkSourceReady(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, pgReadyTimeout)
	defer cancel()

	source, err := s.resolve(ctx, s.payload.GetSourceInstance())
	if err != nil {
		return rerrors.Wrap(err)
	}

	db, err := openPgReplicationDb(source)
	if err != nil {
		return rerrors.Wrap(err)
	}

	defer common.CloseWithLog(db.Close, "pg replication source")

	settings, err := readPgReplicationSettings(ctx, db)
	if err != nil {
		return rerrors.Wrap(err)
	}

	if len(settings.changes()) != 0 {
		return rerrors.Wrap(user_errors.ErrPgReplicationSourceNotReady)
	}

	var inRecovery, hasExternalTablespaces bool

	err = db.QueryRowContext(ctx, `
SELECT pg_is_in_recovery(), EXISTS (
    SELECT 1 FROM pg_tablespace WHERE spcname NOT IN ('pg_default', 'pg_global')
)`).Scan(&inRecovery, &hasExternalTablespaces)
	if err != nil {
		return rerrors.Wrap(err, "error checking source recovery state and tablespaces")
	}

	if inRecovery {
		return rerrors.Wrap(user_errors.ErrPgReplicationSourceInRecovery)
	}

	if hasExternalTablespaces {
		return rerrors.Wrap(user_errors.ErrPgReplicationTablespaces)
	}

	err = checkPgReplicationVersion(ctx, db)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return nil
}

func (s *pgReplicationSetup) checkTargetEmpty(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, pgReadyTimeout)
	defer cancel()

	targetId, err := FindPgInstanceContainerId(ctx, s.docker, s.payload.GetTargetInstance())
	if err != nil {
		return rerrors.Wrap(err)
	}

	if targetId != s.payload.GetTargetContainerId() {
		return rerrors.Wrap(user_errors.ErrPgReplicationVolume)
	}

	replicaSource, err := readPgReplicaMarker(ctx, s.docker.Client(), targetId)
	if err != nil {
		return rerrors.Wrap(err)
	}

	if replicaSource != "" {
		return rerrors.Wrap(user_errors.ErrPgReplicationTargetNotEmpty)
	}

	target, err := s.resolve(ctx, s.payload.GetTargetInstance())
	if err != nil {
		return rerrors.Wrap(err)
	}

	db, err := openPgReplicationDb(target)
	if err != nil {
		return rerrors.Wrap(err)
	}

	defer common.CloseWithLog(db.Close, "pg replication target")

	var inRecovery, hasUserTables bool

	err = db.QueryRowContext(ctx, pgTargetEmptyQuery).Scan(&inRecovery, &hasUserTables)
	if err != nil {
		return rerrors.Wrap(err, "error checking target database")
	}

	if inRecovery || hasUserTables {
		return rerrors.Wrap(user_errors.ErrPgReplicationTargetNotEmpty)
	}

	err = checkPgReplicationVersion(ctx, db)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return nil
}

func (s *pgReplicationSetup) prepareSource(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, pgReadyTimeout)
	defer cancel()

	topology, err := s.backupTopology(ctx)
	if err != nil {
		return rerrors.Wrap(err)
	}

	source, err := s.resolve(ctx, s.payload.GetSourceInstance())
	if err != nil {
		return rerrors.Wrap(err)
	}

	db, err := openPgReplicationDb(source)
	if err != nil {
		return rerrors.Wrap(err)
	}

	defer common.CloseWithLog(db.Close, "pg replication source")

	err = s.ensureReplicationAuth(ctx, db, topology, source.user)
	if err != nil {
		return rerrors.Wrap(err)
	}

	if !s.payload.GetUseSlot() {
		return nil
	}

	slot := s.slotName()

	s.payload.SlotName = slot

	var isUnavailable bool

	err = db.QueryRowContext(ctx, `
SELECT COALESCE((
    SELECT active OR slot_type <> 'physical' OR wal_status = 'lost'
    FROM pg_replication_slots
    WHERE slot_name = $1
), false)`, slot).Scan(&isUnavailable)
	if err != nil {
		return rerrors.Wrap(err, "error checking replication slot")
	}

	if isUnavailable {
		return rerrors.Wrap(user_errors.ErrPgReplicationSlotUnavailable)
	}

	result, err := db.ExecContext(ctx, `
SELECT pg_create_physical_replication_slot($1, true)
WHERE NOT EXISTS (SELECT 1 FROM pg_replication_slots WHERE slot_name = $1)`, slot)
	if err != nil {
		return rerrors.Wrap(err, "error creating physical replication slot")
	}

	created, err := result.RowsAffected()
	if err != nil {
		return rerrors.Wrap(err, "error checking whether a new replication slot was created")
	}

	s.payload.SlotCreated = s.payload.GetSlotCreated() || created > 0

	return nil
}

func (s *pgReplicationSetup) stopTarget(ctx context.Context) error {
	targetId := s.payload.GetTargetContainerId()

	isRunning, exists, err := s.docker.IsContainerRunning(ctx, targetId)
	if err != nil {
		return rerrors.Wrap(err)
	}

	if !exists {
		return rerrors.Wrap(user_errors.ErrContainerNotFoundInEnvironment)
	}

	if !isRunning {
		return nil
	}

	// Check again just before shutdown since the user could have used the target since validation.
	err = s.checkTargetEmpty(ctx)
	if err != nil {
		return rerrors.Wrap(err)
	}

	err = s.docker.Stop(ctx, targetId)
	if err != nil {
		return rerrors.Wrap(err, "error stopping replication target")
	}

	return nil
}

func (s *pgReplicationSetup) startTarget(ctx context.Context) error {
	targetId := s.payload.GetTargetContainerId()
	opts := container.StartOptions{}

	err := s.docker.Client().ContainerStart(ctx, targetId, opts)
	if err != nil {
		return rerrors.Wrap(err, "error starting replication target")
	}

	err = waitForPgReady(ctx, s.docker.Client(), targetId)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return nil
}

func (s *pgReplicationSetup) verifyStreaming(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, pgStreamingTimeout)
	defer cancel()

	source, err := s.resolve(ctx, s.payload.GetSourceInstance())
	if err != nil {
		return rerrors.Wrap(err)
	}

	target := source

	target.host = s.payload.GetTargetInstance()

	targetDb, err := openPgReplicationDb(target)
	if err != nil {
		return rerrors.Wrap(err)
	}

	defer common.CloseWithLog(targetDb.Close, "pg replication target")

	sourceDb, err := openPgReplicationDb(source)
	if err != nil {
		return rerrors.Wrap(err)
	}

	defer common.CloseWithLog(sourceDb.Close, "pg replication source")

	ticker := time.NewTicker(pgStreamingPollInterval)
	defer ticker.Stop()

	for {
		var inRecovery, isStreaming bool

		targetErr := targetDb.QueryRowContext(ctx, "SELECT pg_is_in_recovery()").Scan(&inRecovery)
		sourceErr := sourceDb.QueryRowContext(ctx, `
SELECT EXISTS (
    SELECT 1 FROM pg_stat_replication
    WHERE application_name = $1 AND state = 'streaming'
)`, pgReplicaApplicationName(target.host)).Scan(&isStreaming)

		if targetErr == nil && sourceErr == nil && inRecovery && isStreaming {
			return nil
		}

		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.Canceled) {
				return rerrors.Wrap(ctx.Err())
			}

			return rerrors.Wrap(user_errors.ErrPgReplicationNotStreaming)
		case <-ticker.C:
		}
	}
}

func (s *pgReplicationSetup) slotName() string {
	if name := s.payload.GetSlotName(); name != "" {
		return name
	}

	// DONE can be written before the task context. The same slot is derivable on resume.
	return pgReplicaSlotName(s.payload.GetTargetInstance())
}

func (s *pgReplicationSetup) resolve(ctx context.Context, name string) (replicationEndpoint, error) {
	endpoint, err := resolvePgReplicationInstance(
		ctx, s.dataStorage.Services(), s.dataStorage.PgInstances(), s.secrets, name,
	)
	if err != nil {
		return replicationEndpoint{}, rerrors.Wrap(err)
	}

	return endpoint, nil
}

func openPgReplicationDb(endpoint replicationEndpoint) (*sql.DB, error) {
	db, err := sql.Open(pgReplicationDriver, endpoint.dsn(endpoint.dbName))
	if err != nil {
		return nil, rerrors.Wrap(err, "error opening pg replication connection")
	}

	return db, nil
}

func checkPgReplicationVersion(ctx context.Context, db *sql.DB) error {
	var version int

	err := db.QueryRowContext(ctx, "SHOW server_version_num").Scan(&version)
	if err != nil {
		return rerrors.Wrap(err, "error reading pg server version")
	}

	if version/pgServerVersionMajorDivisor != pgReplicationMajorVersion {
		return rerrors.Wrap(user_errors.ErrPgReplicationVersion)
	}

	return nil
}

func pgReplicaApplicationName(target string) string {
	if len(target) <= pgReplicaSlotMaxLength {
		return target
	}

	return pgReplicaSlotName(target)
}

func pgReplicaSlotName(target string) string {
	const prefix = "velez_"

	sanitized := strings.Map(func(char rune) rune {
		if char >= 'a' && char <= 'z' || char >= '0' && char <= '9' || char == '_' {
			return char
		}

		return '_'
	}, strings.ToLower(target))

	if sanitized == target && len(prefix)+len(sanitized) <= pgReplicaSlotMaxLength {
		return prefix + sanitized
	}

	sum := sha256.Sum256([]byte(target))
	hash := hex.EncodeToString(sum[:pgReplicaSlotHashBytes])
	budget := pgReplicaSlotMaxLength - len(prefix) - len(hash) - len("_")

	return prefix + sanitized[:min(len(sanitized), budget)] + "_" + hash
}

type pgReplicationPrepareSourceJob struct {
	setup *pgReplicationSetup
}

func (j *pgReplicationPrepareSourceJob) Do(ctx context.Context) error {
	err := j.setup.prepareSource(ctx)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return nil
}

func (j *pgReplicationPrepareSourceJob) Rollback(ctx context.Context) error {
	if !j.setup.payload.GetSlotCreated() {
		return nil
	}

	targetIntact, err := j.setup.canRestoreOriginalTarget(ctx)
	if err != nil {
		return rerrors.Wrap(err)
	}

	if !targetIntact {
		return nil
	}

	source, err := j.setup.resolve(ctx, j.setup.payload.GetSourceInstance())
	if err != nil {
		return rerrors.Wrap(err)
	}

	db, err := openPgReplicationDb(source)
	if err != nil {
		return rerrors.Wrap(err)
	}

	defer common.CloseWithLog(db.Close, "pg replication source")

	_, err = db.ExecContext(ctx, `
SELECT pg_drop_replication_slot(slot_name)
FROM pg_replication_slots
WHERE slot_name = $1 AND slot_type = 'physical' AND NOT active`, j.setup.slotName())
	if err != nil {
		return rerrors.Wrap(err, "error removing unused replication slot")
	}

	return nil
}

type pgReplicationStopTargetJob struct {
	setup *pgReplicationSetup
}

func (j *pgReplicationStopTargetJob) Do(ctx context.Context) error {
	err := j.setup.stopTarget(ctx)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return nil
}

func (j *pgReplicationStopTargetJob) Rollback(ctx context.Context) error {
	targetIntact, err := j.setup.canRestoreOriginalTarget(ctx)
	if err != nil {
		return rerrors.Wrap(err)
	}

	if !targetIntact {
		return nil
	}

	opts := container.StartOptions{}

	err = j.setup.docker.Client().ContainerStart(ctx, j.setup.payload.GetTargetContainerId(), opts)
	if err != nil {
		return rerrors.Wrap(err, "error restarting original pg target")
	}

	return nil
}

func (s *pgReplicationSetup) canRestoreOriginalTarget(ctx context.Context) (bool, error) {
	cli := s.docker.Client()
	helper, err := cli.ContainerInspect(ctx, pgBackupHelperPrefix+s.payload.GetOperationId())

	if err == nil && helper.State != nil && helper.State.Running {
		return false, nil
	}

	if err != nil && !errdefs.IsNotFound(err) {
		return false, rerrors.Wrap(err, "error checking for an unfinished pg base backup")
	}

	targetId := s.payload.GetTargetContainerId()

	source, err := readPgReplicaMarker(ctx, cli, targetId)
	if err != nil {
		return false, rerrors.Wrap(err)
	}

	if source != "" {
		return false, nil
	}

	return hasOriginalPgData(ctx, cli, targetId)
}
