package pgaas_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/jobs"
)

func (impl *Impl) SetupPgReplication(
	ctx context.Context,
	req *pb.SetupPgReplication_Request,
) (*pb.SetupPgReplication_Response, error) {
	serviceReq := domain.SetupPgReplicationReq{
		SourceInstance: req.GetSourceInstance(),
		TargetInstance: req.GetTargetInstance(),
		UseSlot:        req.GetUseSlot(),
		Mode:           req.GetMode(),
	}

	err := impl.postgresService.SetupPgReplication(ctx, serviceReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error setting up pg replication")
	}

	response := &pb.SetupPgReplication_Response{
		EntityId: serviceReq.TargetInstance,
		Action:   jobs.SetupPgReplicationAction,
	}

	return response, nil
}

func (impl *Impl) GetPgReplicationStatus(
	ctx context.Context,
	req *pb.GetPgReplicationStatus_Request,
) (*pb.GetPgReplicationStatus_Response, error) {
	status, err := impl.postgresService.GetPgReplicationStatus(ctx, req.GetInstance())
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	replicas := make([]*pb.GetPgReplicationStatus_ReplicaStatus, 0, len(status.Replicas))
	for _, replica := range status.Replicas {
		row := &pb.GetPgReplicationStatus_ReplicaStatus{
			ApplicationName:  replica.ApplicationName,
			State:            replica.State,
			LagBytes:         replica.LagBytes,
			ReplayLagSeconds: replica.ReplayLagSeconds,
			InSync:           replica.InSync,
		}

		replicas = append(replicas, row)
	}

	response := &pb.GetPgReplicationStatus_Response{
		IsReplica:  status.IsReplica,
		Replicas:   replicas,
		Upstream:   status.Upstream,
		InRecovery: status.InRecovery,
		InSync:     status.InSync,
	}

	return response, nil
}
