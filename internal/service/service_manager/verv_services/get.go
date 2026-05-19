package verv_services

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
)

func (v *VervService) Get(ctx context.Context, r domain.GetServiceReq) (domain.Service, error) {
	if r.Name == "" {
		return domain.Service{}, rerrors.New("name is required to find service")
	}

	service, err := v.servicesStorage.GetByName(ctx, r.Name)
	if err != nil {
		return domain.Service{}, rerrors.Wrap(err, "error getting service by name from storage")
	}

	return service, nil
}
