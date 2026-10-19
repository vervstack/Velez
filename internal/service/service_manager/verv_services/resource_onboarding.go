package verv_services

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

// attachOnboarding marks every resource backed by a Velez-managed container. A resource is looked up by its own
// name and by the `<service>_<name>` container name that single-node bindings use.
func (v *VervServicesServiceImpl) attachOnboarding(
	ctx context.Context,
	serviceName string,
	resources []domain.BoundResource,
) error {
	containers, err := v.listAllContainers(ctx)
	if err != nil {
		return rerrors.Wrap(err, "error listing containers for resource onboarding")
	}

	for i := range resources {
		candidates := []string{resources[i].Name, serviceName + "_" + resources[i].Name}

		for _, candidate := range candidates {
			managed, isFound := findManagedContainer(containers, candidate)
			if !isFound {
				continue
			}

			resources[i].IsOnboarded = true
			resources[i].ContainerId = managed.ID

			owner := managed.Labels[labels.VervServiceLabel]
			if owner != serviceName {
				resources[i].LinkedServiceName = owner
			}

			break
		}
	}

	return nil
}
