package verv_services

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
)

func (v *VervService) GetServiceGraph(ctx context.Context, serviceName string) (domain.ServiceGraph, error) {
	deps, err := v.dataStorage.ServiceDependencies().GetDependencies(ctx, serviceName)
	if err != nil {
		return domain.ServiceGraph{}, rerrors.Wrap(err, "error getting service dependencies")
	}

	callers, err := v.dataStorage.ServiceDependencies().GetCallers(ctx, serviceName)
	if err != nil {
		return domain.ServiceGraph{}, rerrors.Wrap(err, "error getting service callers via dependency graph")
	}

	return domain.ServiceGraph{
		Callers:      callers,
		Dependencies: deps,
	}, nil
}
