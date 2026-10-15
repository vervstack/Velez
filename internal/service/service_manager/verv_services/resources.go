package verv_services

import (
	"context"
	"slices"

	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
)

func (v *VervService) GetServiceResources(ctx context.Context, serviceName string) ([]domain.BoundResource, error) {
	resources, err := v.dataStorage.ServiceResources().GetResources(ctx, serviceName)
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting service resources info")
	}

	v.attachWebUis(ctx, serviceName, resources)

	links, err := v.runnerLinks(ctx, serviceName)
	if err != nil {
		log.Ctx(ctx).Warn().
			Str("service", serviceName).
			Err(err).
			Msg("error resolving runner links")
	}

	for _, link := range links {
		isKnown := slices.ContainsFunc(resources, func(resource domain.BoundResource) bool {
			return resource.Name == link.Name
		})
		if isKnown {
			continue
		}

		resource := domain.BoundResource{
			Name:         link.Name,
			ResourceType: link.ResourceType,
			Status:       link.Status,
		}

		resources = append(resources, resource)
	}

	err = v.attachOnboarding(ctx, serviceName, resources)
	if err != nil {
		log.Ctx(ctx).Warn().
			Str("service", serviceName).
			Err(err).
			Msg("error resolving resource onboarding")
	}

	return resources, nil
}
