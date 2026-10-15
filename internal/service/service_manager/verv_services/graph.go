package verv_services

import (
	"context"
	"slices"

	"github.com/rs/zerolog/log"
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

	links, err := v.runnerLinks(ctx, serviceName)
	if err != nil {
		log.Ctx(ctx).Warn().
			Str("service", serviceName).
			Err(err).
			Msg("error resolving runner links")
	}

	for _, link := range links {
		isKnown := slices.ContainsFunc(deps, func(dep domain.ServiceDependency) bool {
			return dep.TargetService == link.Name
		})
		if isKnown {
			continue
		}

		dep := domain.ServiceDependency{
			SourceService: serviceName,
			TargetService: link.Name,
			NodeType:      domain.NodeTypeService,
		}
		if !link.IsManaged {
			dep.NodeType = domain.NodeTypeResource
		}

		deps = append(deps, dep)
	}

	runners, err := v.dindCallers(ctx, serviceName)
	if err != nil {
		log.Ctx(ctx).Warn().
			Str("service", serviceName).
			Err(err).
			Msg("error resolving runner links")
	}

	for _, runner := range runners {
		isKnown := slices.ContainsFunc(callers, func(caller domain.ServiceDependency) bool {
			return caller.SourceService == runner
		})
		if isKnown {
			continue
		}

		caller := domain.ServiceDependency{
			SourceService: runner,
			TargetService: serviceName,
			NodeType:      domain.NodeTypeService,
		}

		callers = append(callers, caller)
	}

	return domain.ServiceGraph{
		Callers:      callers,
		Dependencies: deps,
	}, nil
}
