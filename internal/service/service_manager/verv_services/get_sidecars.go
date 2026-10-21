package verv_services

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/container"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils/network_owner"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/storage/environments"
)

const (
	listAllContainersNoLimit = 0
)

func (v *VervServicesServiceImpl) listSidecars(
	ctx context.Context, serviceName string,
) ([]domain.ServiceSidecar, error) {
	runtime, err := v.runtimes.Runtime(ctx, "")
	if err != nil {
		return nil, rerrors.Wrap(err, "error resolving environment")
	}

	containers, err := runtime.ListAllContainers(ctx, listAllContainersNoLimit)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing containers")
	}

	boundRoots, err := v.boundRootNames(ctx, serviceName)
	if err != nil {
		return nil, rerrors.Wrap(err, "error resolving container bindings")
	}

	var sidecars []domain.ServiceSidecar

	seen := make(map[string]struct{})
	owners := network_owner.OwnerIds(containers)

	for _, root := range containers {
		if !isServiceRoot(root, serviceName, boundRoots) {
			continue
		}

		for _, sidecar := range network_owner.SidecarsOf(containers, root.ID) {
			seen[sidecar.ID] = struct{}{}
			sidecars = append(sidecars, toServiceSidecar(sidecar))
		}

		for _, member := range joinedGroupMembers(containers, owners, root.ID) {
			_, isKnown := seen[member.ID]
			if isKnown {
				continue
			}

			seen[member.ID] = struct{}{}
			sidecars = append(sidecars, toServiceSidecar(member))
		}
	}

	return webUiSidecars(containers, serviceName, sidecars), nil
}

// joinedGroupMembers are the other members of the network group a service container joined:
// the container owning the namespace and everything else sharing it. Empty for a container
// that owns its own namespace.
func joinedGroupMembers(
	containers []container.Summary, owners map[string]string, serviceContainerId string,
) []container.Summary {
	ownerId, isJoined := owners[serviceContainerId]
	if !isJoined {
		return nil
	}

	var members []container.Summary

	for _, cont := range containers {
		if cont.ID == serviceContainerId {
			continue
		}

		if cont.ID == ownerId || owners[cont.ID] == ownerId {
			members = append(members, cont)
		}
	}

	return members
}

// boundRootNames are the containers a binding row ties to the service as its
// own, not as a sidecar. Empty when no bindings backend is live.
func (v *VervServicesServiceImpl) boundRootNames(ctx context.Context, serviceName string) (map[string]struct{}, error) {
	names := make(map[string]struct{})

	bindings := v.dataStorage.ContainerBindings()
	if bindings == nil {
		return names, nil
	}

	list, err := bindings.ListByNode(ctx, domain.SelfNodeId, environments.DefaultEnvironmentName)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing container bindings")
	}

	for _, binding := range list {
		if binding.ServiceName == serviceName && !binding.IsSidecar {
			names[binding.ContainerName] = struct{}{}
		}
	}

	return names, nil
}

// isServiceRoot reports whether the container is a service's own container:
// linked by the service label or a binding row, and not itself a sidecar.
func isServiceRoot(cont container.Summary, serviceName string, boundRoots map[string]struct{}) bool {
	_, isSidecar := cont.Labels[labels.Sidecar]
	if isSidecar {
		return false
	}

	if cont.Labels[labels.VervServiceLabel] == serviceName {
		return true
	}

	_, isBound := boundRoots[summaryName(cont)]

	return isBound
}

func summaryName(summary container.Summary) string {
	if len(summary.Names) == 0 {
		return ""
	}

	return strings.TrimPrefix(summary.Names[0], "/")
}

func toServiceSidecar(summary container.Summary) domain.ServiceSidecar {
	return domain.ServiceSidecar{
		ContainerId:   summary.ID,
		ContainerName: summaryName(summary),
		ImageName:     summary.Image,
		Status:        velez_api.Smerd_Status(velez_api.Smerd_Status_value[summary.State]),
	}
}
