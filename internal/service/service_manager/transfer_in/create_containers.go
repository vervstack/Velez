package transfer_in

import (
	"context"

	"go.redsock.ru/rerrors"
)

// createContainers creates (without starting) the manifest's containers:
// roots first, then the sidecars that join their network namespaces.
func (s *TransferInServiceImpl) createContainers(ctx context.Context, run *receiveRun) error {
	err := ensureNetworks(ctx, run.runtime, run.manifest)
	if err != nil {
		return rerrors.Wrap(err)
	}

	rootIds := make(map[string]string)

	for _, root := range run.manifest.roots() {
		response, createErr := run.runtime.ContainerCreate(ctx, rootCreateRequest(root))
		if createErr != nil {
			return rerrors.Wrap(createErr, "error creating root container")
		}

		run.createdContainerIds = append(run.createdContainerIds, response.ID)
		rootIds[root.name] = response.ID
	}

	for _, sidecar := range run.manifest.sidecars() {
		rootId, isFound := rootIds[sidecar.networkOwner]
		if !isFound {
			return rerrors.Wrap(errRootNotCreated, sidecar.name)
		}

		response, createErr := run.runtime.ContainerCreate(ctx, sidecarCreateRequest(sidecar, rootId))
		if createErr != nil {
			return rerrors.Wrap(createErr, "error creating sidecar container")
		}

		run.createdContainerIds = append(run.createdContainerIds, response.ID)
	}

	return nil
}
