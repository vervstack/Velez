package jobs

import (
	"context"
	"database/sql"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/service/secrets"
	"go.vervstack.ru/Velez/internal/service/user_errors"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/utils/common"
)

const (
	EnablePgReplicationAction = "enable_pg_replication"
	minPgReplicationCapacity  = 10
	pgWalLevelLogical         = "logical"
)

type enablePgReplicationHandler struct {
	dataStorage storage.Storage
	secrets     secrets.Store
	docker      node_clients.Docker
}

func NewEnablePgReplicationHandler(
	dataStorage storage.Storage, secretsStore secrets.Store, docker node_clients.Docker,
) TaskHandler {
	return &enablePgReplicationHandler{dataStorage: dataStorage, secrets: secretsStore, docker: docker}
}

func (h *enablePgReplicationHandler) Action() string { return EnablePgReplicationAction }

func (h *enablePgReplicationHandler) NewContext() TaskContext {
	return &velez_api.EnablePgReplicationTaskPayload{}
}

func (h *enablePgReplicationHandler) BuildJobs(taskCtx TaskContext) []NamedJob {
	payload, ok := taskCtx.(*velez_api.EnablePgReplicationTaskPayload)
	if !ok {
		panic("enable_pg_replication: BuildJobs called with mismatched TaskContext type")
	}

	readiness := &pgReplicationReadiness{
		services:    h.dataStorage.Services(),
		pgInstances: h.dataStorage.PgInstances(),
		secrets:     h.secrets,
		docker:      h.docker,
		payload:     payload,
	}

	return []NamedJob{
		{Name: "check_settings", Job: pgReplicationStep(readiness.checkSettings)},
		{Name: "apply_settings", Job: pgReplicationStep(readiness.applySettings)},
		{Name: "restart_pg", Job: pgReplicationStep(readiness.restartPg)},
		{Name: "wait_ready", Job: pgReplicationStep(readiness.waitReady)},
	}
}

type pgReplicationStep func(context.Context) error

func (step pgReplicationStep) Do(ctx context.Context) error {
	err := step(ctx)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return nil
}

type pgReplicationReadiness struct {
	services    storage.ServicesStorage
	pgInstances storage.PgInstancesStorage
	secrets     secrets.Store
	docker      node_clients.Docker
	payload     *velez_api.EnablePgReplicationTaskPayload
}

func (j *pgReplicationReadiness) checkSettings(ctx context.Context) error {
	changes, err := j.readChanges(ctx)
	if err != nil {
		return rerrors.Wrap(err, "error checking pg replication settings")
	}

	j.payload.RestartNeeded = len(changes) != 0
	j.payload.Applied = !j.payload.GetRestartNeeded()

	return nil
}

func (j *pgReplicationReadiness) applySettings(ctx context.Context) error {
	if j.payload.GetApplied() {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, pgReadyTimeout)
	defer cancel()

	db, err := j.openPg(ctx)
	if err != nil {
		return rerrors.Wrap(err)
	}

	defer common.CloseWithLog(db.Close, "pg replication settings database")

	// A worker may resume after writing only some settings. SHOW still reads
	// active settings, so repeating those ALTER SYSTEM writes is safe.
	settings, err := readPgReplicationSettings(ctx, db)
	if err != nil {
		return rerrors.Wrap(err)
	}

	changes := settings.changes()

	for _, statement := range changes {
		_, err = db.ExecContext(ctx, statement)
		if err != nil {
			return rerrors.Wrapf(err, "error executing %s", statement)
		}
	}

	j.payload.Applied = true
	j.payload.RestartNeeded = len(changes) != 0

	return nil
}

func (j *pgReplicationReadiness) restartPg(ctx context.Context) error {
	if !j.payload.GetApplied() && !j.payload.GetRestartNeeded() {
		// A DONE checkpoint can survive a worker crash before the payload write.
		err := j.checkSettings(ctx)
		if err != nil {
			return rerrors.Wrap(err)
		}
	}

	if !j.payload.GetRestartNeeded() {
		return nil
	}

	changes, err := j.readChanges(ctx)
	if err == nil && len(changes) == 0 {
		// Docker can complete the restart before the worker checkpoints it.
		return nil
	}

	containerId, err := j.findContainerId(ctx)
	if err != nil {
		return rerrors.Wrap(err)
	}

	err = j.docker.Restart(ctx, containerId)
	if err != nil {
		return rerrors.Wrap(err, "error restarting pg instance")
	}

	return nil
}

func (j *pgReplicationReadiness) waitReady(ctx context.Context) error {
	if !j.payload.GetRestartNeeded() {
		return nil
	}

	containerId, err := j.findContainerId(ctx)
	if err != nil {
		return rerrors.Wrap(err)
	}

	err = waitForPgReady(ctx, j.docker.Client(), containerId)
	if err != nil {
		return rerrors.Wrap(err)
	}

	changes, err := j.readChanges(ctx)
	if err != nil {
		return rerrors.Wrap(err, "error verifying pg replication settings after restart")
	}

	if len(changes) != 0 {
		return rerrors.Wrapf(user_errors.ErrPgReplicationSettingsInactive, "still required: %v", changes)
	}

	return nil
}

func (j *pgReplicationReadiness) findContainerId(ctx context.Context) (string, error) {
	containerId, err := FindPgInstanceContainerId(ctx, j.docker, j.payload.GetInstanceName())
	if err != nil {
		return "", rerrors.Wrap(err)
	}

	return containerId, nil
}

func (j *pgReplicationReadiness) readChanges(ctx context.Context) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, pgReadyTimeout)
	defer cancel()

	db, err := j.openPg(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	defer common.CloseWithLog(db.Close, "pg replication settings database")

	settings, err := readPgReplicationSettings(ctx, db)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	return settings.changes(), nil
}

func (j *pgReplicationReadiness) openPg(ctx context.Context) (*sql.DB, error) {
	endpoint, err := resolvePgReplicationInstance(
		ctx, j.services, j.pgInstances, j.secrets, j.payload.GetInstanceName(),
	)
	if err != nil {
		return nil, rerrors.Wrap(err, "error resolving pg instance")
	}

	db, err := sql.Open("postgres", endpoint.dsn(endpoint.dbName))
	if err != nil {
		return nil, rerrors.Wrap(err, "error opening pg instance connection")
	}

	return db, nil
}

type pgReplicationSettings struct {
	walLevel            string
	walLogHints         string
	maxWalSenders       int
	maxReplicationSlots int
}

func readPgReplicationSettings(ctx context.Context, db *sql.DB) (pgReplicationSettings, error) {
	var settings pgReplicationSettings

	err := db.QueryRowContext(ctx, "SHOW wal_level").Scan(&settings.walLevel)
	if err != nil {
		return pgReplicationSettings{}, rerrors.Wrap(err, "error reading wal_level")
	}

	err = db.QueryRowContext(ctx, "SHOW wal_log_hints").Scan(&settings.walLogHints)
	if err != nil {
		return pgReplicationSettings{}, rerrors.Wrap(err, "error reading wal_log_hints")
	}

	err = db.QueryRowContext(ctx, "SHOW max_wal_senders").Scan(&settings.maxWalSenders)
	if err != nil {
		return pgReplicationSettings{}, rerrors.Wrap(err, "error reading max_wal_senders")
	}

	err = db.QueryRowContext(ctx, "SHOW max_replication_slots").Scan(&settings.maxReplicationSlots)
	if err != nil {
		return pgReplicationSettings{}, rerrors.Wrap(err, "error reading max_replication_slots")
	}

	return settings, nil
}

func (s pgReplicationSettings) changes() []string {
	var changes []string

	if s.walLevel != "replica" && s.walLevel != pgWalLevelLogical {
		changes = append(changes, "ALTER SYSTEM SET wal_level = 'replica'")
	}

	if s.walLogHints != "on" {
		changes = append(changes, "ALTER SYSTEM SET wal_log_hints = on")
	}

	if s.maxWalSenders < minPgReplicationCapacity {
		changes = append(changes, "ALTER SYSTEM SET max_wal_senders = 10")
	}

	if s.maxReplicationSlots < minPgReplicationCapacity {
		changes = append(changes, "ALTER SYSTEM SET max_replication_slots = 10")
	}

	return changes
}
