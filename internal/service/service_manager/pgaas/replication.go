package pgaas

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/service/user_errors"
)

func (s *PostgresServiceImpl) SetupPgReplication(ctx context.Context, req domain.SetupPgReplicationReq) error {
	if req.Mode != jobs.PgReplicationModePhysical {
		return rerrors.Wrap(user_errors.ErrPgReplicationMode)
	}

	if req.SourceInstance == "" || req.TargetInstance == "" {
		return rerrors.Wrap(user_errors.ErrInvalidInstanceName)
	}

	if req.SourceInstance == req.TargetInstance {
		return rerrors.Wrap(user_errors.ErrPgReplicationSameInstance)
	}

	_, err := s.getPgInstance(ctx, req.SourceInstance)
	if err != nil {
		return rerrors.Wrap(err, "error validating source pg instance")
	}

	_, err = s.getPgInstance(ctx, req.TargetInstance)
	if err != nil {
		return rerrors.Wrap(err, "error validating target pg instance")
	}

	targetContainerId, err := jobs.FindPgInstanceContainerId(ctx, s.docker, req.TargetInstance)
	if err != nil {
		return rerrors.Wrap(err, "error selecting target pg instance container")
	}

	payload := &velez_api.SetupPgReplicationTaskPayload{
		SourceInstance:    req.SourceInstance,
		TargetInstance:    req.TargetInstance,
		UseSlot:           req.UseSlot,
		Mode:              req.Mode,
		OperationId:       uuid.NewString(),
		TargetContainerId: targetContainerId,
	}

	task, err := s.jobsEngine.EnqueueReplacing(ctx, req.TargetInstance, jobs.SetupPgReplicationAction, payload)
	if err != nil {
		return rerrors.Wrap(err, "error enqueuing setup pg replication task")
	}

	var queued velez_api.SetupPgReplicationTaskPayload

	err = json.Unmarshal(task.Context.RawMessage, &queued)
	if err != nil {
		return rerrors.Wrap(err, "error reading queued pg replication request")
	}

	if queued.GetSourceInstance() != req.SourceInstance || queued.GetUseSlot() != req.UseSlot ||
		queued.GetTargetContainerId() != targetContainerId {
		return rerrors.Wrap(user_errors.ErrPgReplicationAlreadyPending)
	}

	return nil
}

func (s *PostgresServiceImpl) GetPgReplicationStatus(
	ctx context.Context, name string,
) (domain.PgReplicationStatus, error) {
	if name == "" {
		return domain.PgReplicationStatus{}, rerrors.Wrap(user_errors.ErrInvalidInstanceName)
	}

	status, err := jobs.ReadPgReplicationStatus(ctx, s.dataStorage, s.secrets, s.docker, name)
	if err != nil {
		return domain.PgReplicationStatus{}, rerrors.Wrap(err, "error reading pg replication status")
	}

	return status, nil
}

func (s *PostgresServiceImpl) getPgInstance(ctx context.Context, name string) (domain.PgInstance, error) {
	svc, err := s.dataStorage.Services().GetByName(ctx, name)
	if err != nil {
		return domain.PgInstance{}, rerrors.Wrap(err, "error getting pg instance service")
	}

	instance, err := s.dataStorage.PgInstances().GetPgInstanceByServiceID(ctx, svc.ID)
	if err != nil {
		return domain.PgInstance{}, rerrors.Wrap(err, "error getting pg instance row")
	}

	return instance, nil
}

func (s *PostgresServiceImpl) ChangePgMaster(ctx context.Context, req domain.ChangePgMasterReq) error {
	err := jobs.ValidateChangePgMaster(req)
	if err != nil {
		return rerrors.Wrap(err)
	}

	payload := &velez_api.ChangePgMasterTaskPayload{
		BalancerInstance:  req.BalancerInstance,
		NewMasterInstance: req.NewMasterInstance,
		Strategy:          req.OldMasterStrategy,
		OperationId:       uuid.NewString(),
	}

	task, err := s.jobsEngine.EnqueueReplacing(ctx, req.BalancerInstance, jobs.ChangePgMasterAction, payload)
	if err != nil {
		return rerrors.Wrap(err, "error enqueuing pg failover task")
	}

	var queued velez_api.ChangePgMasterTaskPayload

	err = json.Unmarshal(task.Context.RawMessage, &queued)
	if err != nil {
		return rerrors.Wrap(err, "error reading queued pg failover request")
	}

	if queued.GetNewMasterInstance() != req.NewMasterInstance || queued.GetStrategy() != req.OldMasterStrategy {
		return rerrors.Wrap(user_errors.ErrPgFailoverBusy)
	}

	return nil
}
