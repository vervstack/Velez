package dind_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

func (impl *Impl) CreateDind(
	ctx context.Context,
	req *pb.CreateDind_Request,
) (*pb.CreateDind_Response, error) {
	serviceReq := domain.CreateDindReq{
		Name:            req.GetName(),
		Environment:     req.GetEnvironment(),
		IsSysboxEnabled: req.IsSysboxEnabled == nil || req.GetIsSysboxEnabled(),
	}

	err := impl.dindService.CreateDind(ctx, serviceReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating dind")
	}

	return &pb.CreateDind_Response{}, nil
}
