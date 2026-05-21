package service_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) ListEnvironments(ctx context.Context, _ *pb.ListEnvironments_Request) (*pb.ListEnvironments_Response, error) {
	envs, err := impl.servicesService.ListEnvironments(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	resp := &pb.ListEnvironments_Response{
		Environments: envs,
	}

	return resp, nil
}
