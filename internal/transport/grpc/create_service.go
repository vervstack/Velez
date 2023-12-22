package grpc

import (
	"context"

	"github.com/godverv/Velez/internal/domain"
	"github.com/godverv/Velez/pkg/velez_api"
)

func (a *Api) CreateService(ctx context.Context, req *velez_api.CreateService_Request) (*velez_api.CreateService_Response, error) {
	resp, err := a.containerManager.Up(ctx, domain.CreateContainer{})
	if err != nil {
		return nil, err
	}
	_ = resp
	return &velez_api.CreateService_Response{}, nil
}
