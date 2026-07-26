package service_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"
	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (impl *Impl) GetServiceEnvironments(
	ctx context.Context,
	pbReq *pb.GetServiceEnvironments_Request,
) (*pb.GetServiceEnvironments_Response, error) {
	environments, err := impl.servicesService.GetServiceEnvironments(ctx, pbReq.GetServiceName())
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	pbEnvironments := make([]*pb.ServiceEnvironmentInfo, 0, len(environments))
	for _, env := range environments {
		pbEnv := &pb.ServiceEnvironmentInfo{
			Env:             env.Env,
			Status:          env.Status,
			DeployedVersion: env.DeployedVersion,
			Health:          env.Health,
		}

		if env.DeployedAt != nil {
			pbEnv.DeployedAt = timestamppb.New(*env.DeployedAt)
		}

		pbEnvironments = append(pbEnvironments, pbEnv)
	}

	resp := &pb.GetServiceEnvironments_Response{
		Environments: pbEnvironments,
	}

	return resp, nil
}
