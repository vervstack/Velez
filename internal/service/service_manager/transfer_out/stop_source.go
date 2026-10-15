package transfer_out

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
)

func stopSource(ctx context.Context, runtime container_runtime.ContainerRuntime, order sourceOrder) error {
	for _, name := range order.stop {
		err := runtime.Stop(ctx, name)
		if err != nil {
			return rerrors.Wrap(err, "error stopping source container")
		}
	}

	return nil
}
