package service_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/api/grpc"
	"go.vervstack.ru/Velez/internal/domain"
)

func (impl *Impl) CreateService(ctx context.Context, apiReq *pb.CreateService_Request) (*pb.CreateService_Response, error) {
	pipelineReq := domain.CreateServiceReq{
		Name: apiReq.Name,
	}

	runner := impl.pipeliner.CreateService(pipelineReq)
	err := runner.Run(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	return &pb.CreateService_Response{}, nil
}
