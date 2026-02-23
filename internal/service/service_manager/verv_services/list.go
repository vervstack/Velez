package verv_services

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
)

func (v *VervService) List(ctx context.Context, req domain.ListServicesReq) (domain.ServiceList, error) {
	list, err := v.servicesStorage.List(ctx, req)
	if err != nil {
		return domain.ServiceList{}, rerrors.Wrap(err)
	}

	return list, nil
}
