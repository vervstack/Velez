package verv_services

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
)

func (v *VervService) GetServiceGraph(ctx context.Context, serviceName string) (domain.ServiceGraph, error) {
	deps, err := v.serviceDepsStorage.GetDependencies(ctx, serviceName)
	if err != nil {
		return domain.ServiceGraph{}, rerrors.Wrap(err)
	}

	callers, err := v.serviceDepsStorage.GetCallers(ctx, serviceName)
	if err != nil {
		return domain.ServiceGraph{}, rerrors.Wrap(err)
	}

	return domain.ServiceGraph{
		Callers:      callers,
		Dependencies: deps,
	}, nil
}
