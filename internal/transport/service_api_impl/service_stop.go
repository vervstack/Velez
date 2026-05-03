package service_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) StopService(ctx context.Context, req *velez_api.StopService_Request) (*velez_api.StopService_Response, error) {
	err := impl.servicesService.StopService(ctx, req.GetName())
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	return &velez_api.StopService_Response{}, nil
}

func (impl *Impl) RestartService(ctx context.Context, req *velez_api.RestartService_Request) (*velez_api.RestartService_Response, error) {
	err := impl.servicesService.RestartService(ctx, req.GetName())
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	return &velez_api.RestartService_Response{}, nil
}
