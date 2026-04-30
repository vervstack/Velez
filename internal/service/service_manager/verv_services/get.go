package verv_services

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
)

func (v *VervService) Get(ctx context.Context, r domain.GetServiceReq) (domain.Service, error) {
	if r.Id != nil {
		return v.getById(ctx, *r.Id)
	}

	if r.Name != nil {
		return v.getByName(ctx, *r.Name)
	}

	return domain.Service{}, rerrors.New("id or name is required to find service")
}

func (v *VervService) getById(ctx context.Context, id uint64) (domain.Service, error) {
	service, err := v.servicesStorage.GetById(ctx, int64(id))
	if err != nil {
		return domain.Service{}, rerrors.Wrap(err, "error getting service by id from storage")
	}

	return service, nil
}

func (v *VervService) getByName(ctx context.Context, name string) (domain.Service, error) {
	service, err := v.servicesStorage.GetByName(ctx, name)
	if err != nil {
		return domain.Service{}, rerrors.Wrap(err, "error getting service by id from storage")
	}

	return service, nil
}
