package service_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

func (impl *Impl) RemoveService(
	ctx context.Context,
	req *velez_api.RemoveService_Request,
) (*velez_api.RemoveService_Response, error) {
	removeReq := domain.RemoveServiceReq{
		Name:                 req.GetName(),
		DropRunningInstances: req.GetDropRunningInstances(),
	}

	err := impl.servicesService.Remove(ctx, removeReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error removing service")
	}

	return &velez_api.RemoveService_Response{}, nil
}
