package transfer_out

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils/network_owner"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	listAllContainersNoLimit = 0
)

// resolved is the live picture of a container and everything bound to it.
type resolved struct {
	runtime   container_runtime.ContainerRuntime
	root      container.InspectResponse
	sidecars  []container.InspectResponse
	consumers []domain.TransferConsumer
}

func (s *Service) resolve(ctx context.Context, containerName string) (resolved, error) {
	runtime, err := s.runtimes.Runtime(ctx, "")
	if err != nil {
		return resolved{}, rerrors.Wrap(err, "error resolving environment")
	}

	root, err := inspectRoot(ctx, runtime, containerName)
	if err != nil {
		return resolved{}, rerrors.Wrap(err)
	}

	sidecars, err := inspectSidecars(ctx, runtime, root)
	if err != nil {
		return resolved{}, rerrors.Wrap(err)
	}

	consumers, err := s.findConsumers(ctx, root, sidecars)
	if err != nil {
		return resolved{}, rerrors.Wrap(err)
	}

	res := resolved{
		runtime:   runtime,
		root:      root,
		sidecars:  sidecars,
		consumers: consumers,
	}

	return res, nil
}

func inspectRoot(
	ctx context.Context,
	runtime container_runtime.ContainerRuntime,
	containerName string,
) (container.InspectResponse, error) {
	root, isFound, err := runtime.Inspect(ctx, containerName)
	if err != nil {
		return container.InspectResponse{}, rerrors.Wrap(err, "error inspecting container")
	}

	if !isFound {
		return container.InspectResponse{}, rerrors.Wrap(errContainerNotFound)
	}

	isSidecar := root.Config != nil && hasSidecarLabel(root.Config.Labels)
	if isSidecar {
		return container.InspectResponse{}, rerrors.Wrap(errSidecarNotMovable)
	}

	return root, nil
}

func hasSidecarLabel(containerLabels map[string]string) bool {
	_, isSidecar := containerLabels[labels.Sidecar]

	return isSidecar
}

func inspectSidecars(
	ctx context.Context,
	runtime container_runtime.ContainerRuntime,
	root container.InspectResponse,
) ([]container.InspectResponse, error) {
	all, err := runtime.ListAllContainers(ctx, listAllContainersNoLimit)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing containers")
	}

	summaries := network_owner.SidecarsOf(all, root.ID)
	sidecars := make([]container.InspectResponse, 0, len(summaries))

	for _, summary := range summaries {
		inspected, isFound, inspectErr := runtime.Inspect(ctx, summary.ID)
		if inspectErr != nil {
			return nil, rerrors.Wrap(inspectErr, "error inspecting sidecar")
		}

		if isFound {
			sidecars = append(sidecars, inspected)
		}
	}

	return sidecars, nil
}

func (r resolved) toPlan() domain.TransferPlan {
	sidecars := make([]domain.TransferSidecar, 0, len(r.sidecars))

	for _, sidecar := range r.sidecars {
		item := domain.TransferSidecar{
			ContainerName: containerName(sidecar),
			ImageName:     imageRef(sidecar),
		}

		sidecars = append(sidecars, item)
	}

	return domain.TransferPlan{
		Sidecars:          sidecars,
		Consumers:         r.consumers,
		UnsupportedMounts: bindMountSources(r.root),
	}
}

func imageRef(inspected container.InspectResponse) string {
	if inspected.Config == nil {
		return ""
	}

	return inspected.Config.Image
}
