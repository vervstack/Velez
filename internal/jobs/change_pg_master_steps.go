package jobs

import (
	"context"
	"database/sql"
	"errors"
	"path"
	"time"

	"github.com/docker/docker/api/types/container"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/service/user_errors"
	"go.vervstack.ru/Velez/internal/utils/common"
)

func (j *pgMasterChange) promote(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, pgStreamingTimeout)
	defer cancel()

	err := j.confirmContainers(ctx)
	if err != nil {
		return rerrors.Wrap(err)
	}

	err = j.confirmBalancer(ctx, false)
	if err != nil {
		return rerrors.Wrap(err)
	}

	oldEndpoint, newEndpoint, err := j.endpoints(ctx)
	if err != nil {
		return rerrors.Wrap(err)
	}

	newDb, err := openPgReplicationDb(newEndpoint)
	if err != nil {
		return rerrors.Wrap(err)
	}

	defer common.CloseWithLog(newDb.Close, "new pg master")

	var isReplica bool

	err = newDb.QueryRowContext(ctx, "SELECT pg_is_in_recovery()").Scan(&isReplica)
	if err != nil {
		return rerrors.Wrap(err, "error checking promotion state")
	}

	if isReplica {
		err = j.requestPromotion(ctx, oldEndpoint, newEndpoint, newDb)
		if err != nil {
			return rerrors.Wrap(err)
		}
	}

	err = waitForPgPromotion(ctx, newDb)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return j.clearPromotedReplicaMarker(ctx)
}

func (j *pgMasterChange) requestPromotion(
	ctx context.Context, oldEndpoint, newEndpoint replicationEndpoint, newDb *sql.DB,
) error {
	oldDb, err := openPgReplicationDb(oldEndpoint)
	if err != nil {
		return rerrors.Wrap(err)
	}

	defer common.CloseWithLog(oldDb.Close, "old pg master")

	err = j.warnAboutClientSessions(ctx, oldDb)
	if err != nil {
		return rerrors.Wrap(err)
	}

	err = j.savePrimaryCredentials(ctx, newEndpoint)
	if err != nil {
		return rerrors.Wrap(err)
	}

	err = j.requireInSync(ctx, oldDb, newDb)
	if err != nil {
		return rerrors.Wrap(err)
	}

	var isPromotionRequested bool

	err = newDb.QueryRowContext(ctx, "SELECT pg_promote(false)").Scan(&isPromotionRequested)
	if err != nil {
		return rerrors.Wrap(err, "error requesting PostgreSQL promotion")
	}

	if !isPromotionRequested {
		return rerrors.Wrap(user_errors.ErrPgFailoverPromoteTimeout)
	}

	return nil
}

func waitForPgPromotion(ctx context.Context, db *sql.DB) error {
	ticker := time.NewTicker(pgStreamingPollInterval)
	defer ticker.Stop()

	for {
		var isReplica bool

		err := db.QueryRowContext(ctx, "SELECT pg_is_in_recovery()").Scan(&isReplica)
		if err == nil && !isReplica {
			return nil
		}

		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.Canceled) {
				return rerrors.Wrap(ctx.Err())
			}

			return rerrors.Wrap(user_errors.ErrPgFailoverPromoteTimeout)
		case <-ticker.C:
		}
	}
}

func (j *pgMasterChange) clearPromotedReplicaMarker(ctx context.Context) error {
	opts := container.ExecOptions{
		User:         pgSourceProcessUser,
		Cmd:          []string{"rm", "-f", "--", path.Join(pgReplicaDataPath, pgReplicaMarkerFile)},
		AttachStdout: true,
		AttachStderr: true,
	}

	return runPgReplicationExec(ctx, j.handler.docker.Client(), j.payload.GetNewMasterContainerId(), opts)
}

func (j *pgMasterChange) switchBalancer(ctx context.Context) error {
	err := j.confirmContainers(ctx)
	if err != nil {
		return rerrors.Wrap(err)
	}

	err = j.confirmBalancer(ctx, true)
	if err != nil {
		return rerrors.Wrap(err)
	}

	request := &velez_api.SelectHaproxyMaster_Request{
		Name: j.payload.GetBalancerInstance(), MasterPgInstance: j.payload.GetNewMasterInstance(),
	}
	selection := &velez_api.SelectHaproxyMasterTaskPayload{
		Request: request, Environment: j.payload.GetEnvironment(),
		ContainerId: j.payload.GetBalancerContainerId(), Port: j.payload.GetBalancerPort(),
	}
	steps := []Job{
		&stageHaproxyConfigJob{
			payload: selection, configResolver: j.handler.configResolver, runtimes: j.handler.runtimes,
		},
		&checkHaproxyConfigJob{payload: selection, runtimes: j.handler.runtimes},
		&reloadHaproxyJob{payload: selection, runtimes: j.handler.runtimes},
	}

	for _, step := range steps {
		err = step.Do(ctx)
		if err != nil {
			return rerrors.Wrap(err)
		}
	}

	return nil
}

func (j *pgMasterChange) replicationSetup() *pgReplicationSetup {
	payload := &velez_api.SetupPgReplicationTaskPayload{
		SourceInstance:    j.payload.GetNewMasterInstance(),
		TargetInstance:    j.payload.GetOldMasterInstance(),
		TargetContainerId: j.payload.GetOldMasterContainerId(),
		OperationId:       j.payload.GetOperationId(),
		UseSlot:           true,
	}

	return &pgReplicationSetup{
		dataStorage: j.handler.dataStorage, secrets: j.handler.secrets, docker: j.handler.docker, payload: payload,
	}
}

func (j *pgMasterChange) prepareSlot(ctx context.Context) error {
	err := j.confirmContainers(ctx)
	if err != nil {
		return rerrors.Wrap(err)
	}

	setup := j.replicationSetup()

	err = setup.prepareSource(ctx)
	if err != nil {
		return rerrors.Wrap(err)
	}

	if j.payload.GetStrategy() != pgFailoverRewind {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, pgReadyTimeout)
	defer cancel()

	endpoint, err := setup.resolve(ctx, j.payload.GetNewMasterInstance())
	if err != nil {
		return rerrors.Wrap(err)
	}

	db, err := openPgReplicationDb(endpoint)
	if err != nil {
		return rerrors.Wrap(err)
	}

	defer common.CloseWithLog(db.Close, "new pg master")

	// pg_rewind reads the source control file to discover the promoted timeline.
	_, err = db.ExecContext(ctx, "CHECKPOINT")
	if err != nil {
		return rerrors.Wrap(err, "error checkpointing the new pg master timeline")
	}

	return nil
}

func (j *pgMasterChange) stopOldMaster(ctx context.Context) error {
	err := j.confirmContainers(ctx)
	if err != nil {
		return rerrors.Wrap(err)
	}

	err = j.handler.docker.Stop(ctx, j.payload.GetOldMasterContainerId())
	if err != nil {
		return rerrors.Wrap(err, "error stopping the old pg master")
	}

	return nil
}

func (j *pgMasterChange) reintegrate(ctx context.Context) error {
	err := j.confirmContainers(ctx)
	if err != nil {
		return rerrors.Wrap(err)
	}

	setup := j.replicationSetup()

	switch j.payload.GetStrategy() {
	case pgFailoverRewind:
		return setup.rewind(ctx)
	case pgFailoverRecopy:
		return setup.basebackup(ctx)
	default:
		return rerrors.Wrap(user_errors.ErrPgFailoverStrategy)
	}
}

func (j *pgMasterChange) startOldMaster(ctx context.Context) error {
	err := j.confirmContainers(ctx)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return j.replicationSetup().startTarget(ctx)
}

func (j *pgMasterChange) verifyStreaming(ctx context.Context) error {
	return j.replicationSetup().verifyStreaming(ctx)
}
