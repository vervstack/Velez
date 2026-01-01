package service_api_impl

import (
	"context"

	pb "go.vervstack.ru/Velez/pkg/velez_api"
)

func (impl *Impl) CreateDeploy(ctx context.Context, apiReq *pb.CreateDeploy_Request) (
	*pb.CreateDeploy_Response, error) {
	return &pb.CreateDeploy_Response{}, nil
}
