package jobs

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/service/secrets"
	"go.vervstack.ru/Velez/internal/service/user_errors"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/utils/common"
)

const (
	ChangePgMasterAction = "change_pg_master"
	pgFailoverRewind     = "rewind"
	pgFailoverRecopy     = "recopy"
	pgStreamingState     = "streaming"
)

type changePgMasterHandler struct {
	dataStorage    storage.Storage
	secrets        secrets.Store
	docker         node_clients.Docker
	balancers      service.HaproxyService
	configResolver service.ServiceConfigResolver
	runtimes       container_runtime.RuntimeResolver
}

func NewChangePgMasterHandler(
	dataStorage storage.Storage, secretsStore secrets.Store, docker node_clients.Docker,
	balancers service.HaproxyService, configResolver service.ServiceConfigResolver,
	runtimes container_runtime.RuntimeResolver,
) TaskHandler {
	return &changePgMasterHandler{
		dataStorage: dataStorage, secrets: secretsStore, docker: docker, balancers: balancers,
		configResolver: configResolver, runtimes: runtimes,
	}
}

func (h *changePgMasterHandler) Action() string { return ChangePgMasterAction }

func (h *changePgMasterHandler) NewContext() TaskContext {
	return &velez_api.ChangePgMasterTaskPayload{}
}

func (h *changePgMasterHandler) BuildJobs(taskCtx TaskContext) []NamedJob {
	payload, ok := taskCtx.(*velez_api.ChangePgMasterTaskPayload)
	if !ok {
		panic("change_pg_master: BuildJobs called with mismatched TaskContext type")
	}

	failover := &pgMasterChange{handler: h, payload: payload}

	return []NamedJob{
		{Name: "resolve", Job: pgReplicationStep(failover.resolve)},
		{Name: "check_in_sync", Job: pgReplicationStep(failover.checkInSync)},
		{Name: "promote", Job: pgReplicationStep(failover.promote)},
		{Name: "switch_balancer", Job: pgReplicationStep(failover.switchBalancer)},
		{Name: "stop_old_master", Job: pgReplicationStep(failover.stopOldMaster)},
		{Name: "prepare_old_master_slot", Job: pgReplicationStep(failover.prepareSlot)},
		{Name: "reintegrate_old_master", Job: pgReplicationStep(failover.reintegrate)},
		{Name: "start_old_master", Job: pgReplicationStep(failover.startOldMaster)},
		{Name: "verify_streaming", Job: pgReplicationStep(failover.verifyStreaming)},
	}
}

func ValidateChangePgMaster(req domain.ChangePgMasterReq) error {
	switch {
	case req.BalancerInstance == "" || req.NewMasterInstance == "":
		return rerrors.Wrap(user_errors.ErrInvalidInstanceName)
	case req.OldMasterStrategy != pgFailoverRewind && req.OldMasterStrategy != pgFailoverRecopy:
		return rerrors.Wrap(user_errors.ErrPgFailoverStrategy)
	case req.ConfirmOutOfSync:
		return rerrors.Wrap(user_errors.ErrPgFailoverNoOverride)
	default:
		return nil
	}
}

type pgMasterChange struct {
	handler *changePgMasterHandler
	payload *velez_api.ChangePgMasterTaskPayload
}

func (j *pgMasterChange) resolve(ctx context.Context) error {
	req := domain.ChangePgMasterReq{
		BalancerInstance:  j.payload.GetBalancerInstance(),
		NewMasterInstance: j.payload.GetNewMasterInstance(),
		OldMasterStrategy: j.payload.GetStrategy(),
	}

	err := ValidateChangePgMaster(req)
	if err != nil {
		return rerrors.Wrap(err)
	}

	err = uuid.Validate(j.payload.GetOperationId())
	if err != nil {
		return rerrors.Wrap(user_errors.ErrPgReplicationOperationMissing)
	}

	balancer, balancerId, err := j.handler.balancers.ResolveMasterSelection(ctx, req.BalancerInstance)
	if err != nil {
		return rerrors.Wrap(err)
	}

	if balancer.MasterPgInstance == req.NewMasterInstance {
		return rerrors.Wrap(user_errors.ErrPgFailoverUpstream)
	}

	j.payload.OldMasterInstance = balancer.MasterPgInstance
	j.payload.Environment = balancer.Environment
	j.payload.BalancerContainerId = balancerId
	j.payload.BalancerPort = balancer.Port

	err = j.pinPgContainers(ctx)
	if err != nil {
		return rerrors.Wrap(err)
	}

	_, err = j.replicationSetup().backupTopology(ctx)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return j.validatePgRoles(ctx)
}

func (j *pgMasterChange) pinPgContainers(ctx context.Context) error {
	docker := j.handler.docker

	oldId, err := FindPgInstanceContainerId(ctx, docker, j.payload.GetOldMasterInstance())
	if err != nil {
		return rerrors.Wrap(err)
	}

	newId, err := FindPgInstanceContainerId(ctx, docker, j.payload.GetNewMasterInstance())
	if err != nil {
		return rerrors.Wrap(err)
	}

	upstream, err := readPgReplicaMarker(ctx, docker.Client(), newId)
	if err != nil {
		return rerrors.Wrap(err)
	}

	if upstream != j.payload.GetOldMasterInstance() {
		return rerrors.Wrap(user_errors.ErrPgFailoverUpstream)
	}

	j.payload.OldMasterContainerId = oldId
	j.payload.NewMasterContainerId = newId

	return nil
}

func (j *pgMasterChange) validatePgRoles(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, pgReadyTimeout)
	defer cancel()

	oldDb, newDb, err := j.openDatabases(ctx)
	if err != nil {
		return rerrors.Wrap(err)
	}

	defer common.CloseWithLog(oldDb.Close, "old pg master")
	defer common.CloseWithLog(newDb.Close, "new pg master")

	var isOldReplica, hasTablespaces, hasWalLogHints, isNewReplica bool

	err = oldDb.QueryRowContext(ctx, `
SELECT pg_is_in_recovery(), EXISTS (
    SELECT 1 FROM pg_tablespace WHERE spcname NOT IN ('pg_default', 'pg_global')
), current_setting('wal_log_hints')::boolean`).Scan(&isOldReplica, &hasTablespaces, &hasWalLogHints)
	if err != nil {
		return rerrors.Wrap(err, "error checking old pg master")
	}

	err = newDb.QueryRowContext(ctx, "SELECT pg_is_in_recovery()").Scan(&isNewReplica)
	if err != nil {
		return rerrors.Wrap(err, "error checking promotion candidate")
	}

	switch {
	case isOldReplica || !isNewReplica:
		return rerrors.Wrap(user_errors.ErrPgFailoverUpstream)
	case hasTablespaces:
		return rerrors.Wrap(user_errors.ErrPgReplicationTablespaces)
	case j.payload.GetStrategy() == pgFailoverRewind && !hasWalLogHints:
		return rerrors.Wrap(user_errors.ErrPgFailoverRewindHints)
	}

	err = checkPgReplicationVersion(ctx, oldDb)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return checkPgReplicationVersion(ctx, newDb)
}

func (j *pgMasterChange) checkInSync(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, pgReadyTimeout)
	defer cancel()

	oldDb, newDb, err := j.openDatabases(ctx)
	if err != nil {
		return rerrors.Wrap(err)
	}

	defer common.CloseWithLog(oldDb.Close, "old pg master")
	defer common.CloseWithLog(newDb.Close, "new pg master")

	return j.requireInSync(ctx, oldDb, newDb)
}

func (j *pgMasterChange) requireInSync(ctx context.Context, oldDb, newDb *sql.DB) error {
	var upstream, state, replayLsn string

	err := newDb.QueryRowContext(ctx, `
SELECT COALESCE(sender_host, ''), status, COALESCE(pg_last_wal_replay_lsn()::text, '')
FROM pg_stat_wal_receiver
WHERE pg_is_in_recovery()`).Scan(&upstream, &state, &replayLsn)
	if errors.Is(err, sql.ErrNoRows) {
		return rerrors.Wrap(user_errors.ErrPgFailoverOutOfSync)
	}

	if err != nil {
		return rerrors.Wrap(err, "error checking the replica WAL receiver")
	}

	err = checkPgFailoverSync(j.payload.GetOldMasterInstance(), upstream, state, replayLsn != "")
	if err != nil {
		return rerrors.Wrap(err)
	}

	isInSync, err := queryPgStandbyInSync(ctx, oldDb, j.payload.GetNewMasterInstance(), replayLsn)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return checkPgFailoverSync(j.payload.GetOldMasterInstance(), upstream, state, isInSync)
}

func checkPgFailoverSync(expectedUpstream, upstream, state string, isInSync bool) error {
	if state != pgStreamingState {
		return rerrors.Wrap(user_errors.ErrPgFailoverOutOfSync)
	}

	if upstream != expectedUpstream {
		return rerrors.Wrap(user_errors.ErrPgFailoverUpstream)
	}

	if !isInSync {
		return rerrors.Wrap(user_errors.ErrPgFailoverOutOfSync)
	}

	return nil
}

func (j *pgMasterChange) openDatabases(ctx context.Context) (oldDb, newDb *sql.DB, err error) {
	oldEndpoint, newEndpoint, err := j.endpoints(ctx)
	if err != nil {
		return nil, nil, rerrors.Wrap(err)
	}

	oldDb, err = openPgReplicationDb(oldEndpoint)
	if err != nil {
		return nil, nil, rerrors.Wrap(err)
	}

	newDb, err = openPgReplicationDb(newEndpoint)
	if err != nil {
		common.CloseWithLog(oldDb.Close, "old pg master")

		return nil, nil, rerrors.Wrap(err)
	}

	return oldDb, newDb, nil
}

func (j *pgMasterChange) endpoints(ctx context.Context) (oldEndpoint, newEndpoint replicationEndpoint, err error) {
	setup := j.replicationSetup()

	oldEndpoint, err = setup.resolve(ctx, j.payload.GetOldMasterInstance())
	if err != nil {
		return replicationEndpoint{}, replicationEndpoint{}, rerrors.Wrap(err)
	}

	newEndpoint, err = setup.resolve(ctx, j.payload.GetNewMasterInstance())
	if err != nil {
		return replicationEndpoint{}, replicationEndpoint{}, rerrors.Wrap(err)
	}

	newEndpoint.user = oldEndpoint.user
	newEndpoint.password = oldEndpoint.password
	newEndpoint.dbName = oldEndpoint.dbName

	return oldEndpoint, newEndpoint, nil
}

func (j *pgMasterChange) warnAboutClientSessions(ctx context.Context, db *sql.DB) error {
	var connections int

	err := db.QueryRowContext(ctx, `
SELECT count(*) FROM pg_stat_activity WHERE backend_type = 'client backend' AND pid <> pg_backend_pid()
`).Scan(&connections)
	if err != nil {
		return rerrors.Wrap(err, "error checking active PostgreSQL client connections")
	}

	if connections > 0 {
		log.Ctx(ctx).Warn().Str("instance", j.payload.GetOldMasterInstance()).Int("connections", connections).
			Msg("Client connections remain open on the old PostgreSQL master during failover")
	}

	return nil
}

func (j *pgMasterChange) confirmContainers(ctx context.Context) error {
	docker := j.handler.docker

	oldId, err := FindPgInstanceContainerId(ctx, docker, j.payload.GetOldMasterInstance())
	if err != nil {
		return rerrors.Wrap(err)
	}

	newId, err := FindPgInstanceContainerId(ctx, docker, j.payload.GetNewMasterInstance())
	if err != nil {
		return rerrors.Wrap(err)
	}

	if oldId != j.payload.GetOldMasterContainerId() || newId != j.payload.GetNewMasterContainerId() {
		return rerrors.Wrap(user_errors.ErrPgFailoverTopologyChanged)
	}

	return nil
}

func (j *pgMasterChange) confirmBalancer(ctx context.Context, mayBeSwitched bool) error {
	balancer, id, err := j.handler.balancers.ResolveMasterSelection(ctx, j.payload.GetBalancerInstance())
	if err != nil {
		return rerrors.Wrap(err)
	}

	master := balancer.MasterPgInstance
	isExpectedMaster := master == j.payload.GetOldMasterInstance() ||
		mayBeSwitched && master == j.payload.GetNewMasterInstance()

	if id != j.payload.GetBalancerContainerId() || balancer.Environment != j.payload.GetEnvironment() ||
		balancer.Port != j.payload.GetBalancerPort() || !isExpectedMaster {
		return rerrors.Wrap(user_errors.ErrPgFailoverTopologyChanged)
	}

	return nil
}
