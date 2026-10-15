package transfer_out

import (
	"context"

	"github.com/rs/zerolog/log"

	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
)

// restoreSource brings back what was running before the transfer, best effort.
func restoreSource(ctx context.Context, runtime container_runtime.ContainerRuntime, order sourceOrder) {
	for _, name := range order.restore {
		err := runtime.Restart(ctx, name)
		if err != nil {
			log.Ctx(ctx).Error().
				Str("container_name", name).
				Err(err).
				Msg("error restoring source container after failed transfer")
		}
	}
}
