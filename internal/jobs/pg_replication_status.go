package jobs

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service/secrets"
	"go.vervstack.ru/Velez/internal/service/user_errors"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/utils/common"
)

const (
	pgReplicationStatusTimeout = 4 * time.Second
	pgReplicationStatusQuery   = `
WITH current_position AS MATERIALIZED (SELECT pg_current_wal_lsn() AS lsn)
SELECT
    application_name,
    state,
    pg_wal_lsn_diff(current_position.lsn, replay_lsn)::bigint,
    EXTRACT(EPOCH FROM replay_lag)::real,
    COALESCE(replay_lsn = current_position.lsn, false)
FROM pg_stat_replication
CROSS JOIN current_position
ORDER BY application_name, pid`
)

func ReadPgReplicaSource(ctx context.Context, docker node_clients.Docker, name string) (string, error) {
	id, err := FindPgInstanceContainerId(ctx, docker, name)
	if errors.Is(err, user_errors.ErrContainerNotFoundInEnvironment) {
		return "", nil
	}

	if err != nil {
		return "", rerrors.Wrap(err)
	}

	source, err := readPgReplicaMarker(ctx, docker.Client(), id)
	if err != nil {
		return "", rerrors.Wrap(err)
	}

	return source, nil
}

func ReadPgReplicationStatus(
	ctx context.Context, dataStorage storage.Storage, secretsStore secrets.Store, docker node_clients.Docker, name string,
) (domain.PgReplicationStatus, error) {
	ctx, cancel := context.WithTimeout(ctx, pgReplicationStatusTimeout)
	defer cancel()

	services := dataStorage.Services()
	pgInstances := dataStorage.PgInstances()

	endpoint, err := resolvePgReplicationInstance(ctx, services, pgInstances, secretsStore, name)
	if err != nil {
		return domain.PgReplicationStatus{}, rerrors.Wrap(err)
	}

	upstreamName, err := ReadPgReplicaSource(ctx, docker, name)
	if err != nil {
		return domain.PgReplicationStatus{}, rerrors.Wrap(err)
	}

	var upstream replicationEndpoint

	if upstreamName != "" {
		upstream, err = resolvePgReplicationInstance(ctx, services, pgInstances, secretsStore, upstreamName)
		if err != nil {
			return domain.PgReplicationStatus{}, rerrors.Wrap(err, "error resolving replica credentials")
		}

		endpoint.user = upstream.user
		endpoint.password = upstream.password
		endpoint.dbName = upstream.dbName
	}

	db, err := openPgReplicationDb(endpoint)
	if err != nil {
		return domain.PgReplicationStatus{}, rerrors.Wrap(err)
	}

	defer common.CloseWithLog(db.Close, "pg replication status database")

	var inRecovery bool

	err = db.QueryRowContext(ctx, "SELECT pg_is_in_recovery()").Scan(&inRecovery)
	if err != nil {
		return domain.PgReplicationStatus{}, rerrors.Wrap(err, "error reading pg recovery state")
	}

	if inRecovery {
		status, standbyErr := readPgStandbyStatus(ctx, db, upstreamName, upstream, name)
		if standbyErr != nil {
			return domain.PgReplicationStatus{}, rerrors.Wrap(standbyErr)
		}

		return status, nil
	}

	replicas, err := readPgPrimaryReplicas(ctx, db)
	if err != nil {
		return domain.PgReplicationStatus{}, rerrors.Wrap(err)
	}

	status := domain.PgReplicationStatus{Replicas: replicas}

	return status, nil
}

func readPgPrimaryReplicas(ctx context.Context, db *sql.DB) ([]domain.PgReplicaStatus, error) {
	rows, err := db.QueryContext(ctx, pgReplicationStatusQuery)
	if err != nil {
		return nil, rerrors.Wrap(err, "error reading pg replication senders")
	}

	defer common.CloseWithLog(rows.Close, "pg replication sender rows")

	var replicas []domain.PgReplicaStatus

	for rows.Next() {
		var (
			name, state       string
			lagBytes          sql.NullInt64
			replayLag         sql.NullFloat64
			atCurrentPosition bool
		)

		err = rows.Scan(&name, &state, &lagBytes, &replayLag, &atCurrentPosition)
		if err != nil {
			return nil, rerrors.Wrap(err, "error reading pg replication sender")
		}

		replicas = append(replicas, pgReplicaStatus(name, state, lagBytes, replayLag, atCurrentPosition))
	}

	err = rows.Err()
	if err != nil {
		return nil, rerrors.Wrap(err, "error iterating pg replication senders")
	}

	return replicas, nil
}

func readPgStandbyStatus(
	ctx context.Context, db *sql.DB, upstreamName string, upstream replicationEndpoint, name string,
) (domain.PgReplicationStatus, error) {
	status := domain.PgReplicationStatus{IsReplica: true, InRecovery: true, Upstream: upstreamName}

	var host, receiverState string

	err := db.QueryRowContext(ctx, "SELECT sender_host, status FROM pg_stat_wal_receiver").Scan(&host, &receiverState)
	if errors.Is(err, sql.ErrNoRows) {
		return status, nil
	}

	if err != nil {
		return domain.PgReplicationStatus{}, rerrors.Wrap(err, "error reading pg replication receiver")
	}

	status.Upstream = host
	if receiverState == "streaming" && host == upstreamName {
		status.InSync = isPgStandbyInSync(ctx, upstream, name)
	}

	return status, nil
}

func isPgStandbyInSync(ctx context.Context, upstream replicationEndpoint, name string) bool {
	db, err := openPgReplicationDb(upstream)
	if err != nil {
		return false
	}

	defer common.CloseWithLog(db.Close, "pg replication upstream")

	var isInSync bool

	err = db.QueryRowContext(ctx, `
SELECT EXISTS (
    SELECT 1 FROM pg_stat_replication
    WHERE application_name = $1
      AND state = 'streaming'
      AND replay_lsn = pg_current_wal_lsn()
)`, pgReplicaApplicationName(name)).Scan(&isInSync)

	return err == nil && isInSync
}

func pgReplicaStatus(
	name, state string, lagBytes sql.NullInt64, replayLag sql.NullFloat64, atCurrentPosition bool,
) domain.PgReplicaStatus {
	status := domain.PgReplicaStatus{
		ApplicationName:  name,
		State:            state,
		LagBytes:         max(0, lagBytes.Int64),
		ReplayLagSeconds: float32(max(0, replayLag.Float64)),
		InSync:           atCurrentPosition,
	}

	return status
}
