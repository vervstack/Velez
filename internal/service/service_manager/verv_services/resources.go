package verv_services

import (
	"context"

	"go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/domain"
)

func (v *VervService) GetServiceResources(ctx context.Context, serviceName string) ([]domain.BoundResource, error) {
	resources, err := v.serviceResourcesStorage.GetResources(ctx, serviceName)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	return resources, nil
}
