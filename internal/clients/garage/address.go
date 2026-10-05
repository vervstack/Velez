package garage

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain"
)

const (
	httpScheme = "http://"
)

// AdminUrl resolves the address Velez reaches the admin API of the Garage
// container on.
func AdminUrl(
	ctx context.Context,
	runtime container_runtime.ContainerRuntime,
	containerName string,
) (string, error) {
	info, isFound, err := runtime.Inspect(ctx, containerName)
	if err != nil {
		return "", rerrors.Wrap(err, "error inspecting garage container")
	}

	if !isFound {
		return "", rerrors.Wrap(errContainerNotFound)
	}

	address, err := runtime.ContainerAddress(info, domain.S3AdminContainerPort)
	if err != nil {
		return "", rerrors.Wrap(err, "error resolving garage admin address")
	}

	return httpScheme + address, nil
}
