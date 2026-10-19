package transfer_in

import (
	"context"

	"github.com/containerd/errdefs"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
)

// preflight checks the whole manifest against this node before anything is
// created, so a conflict rejects the transfer instead of leaving half of it.
func (s *TransferInServiceImpl) preflight(
	ctx context.Context,
	runtime container_runtime.ContainerRuntime,
	manifest transferManifest,
) error {
	for _, c := range manifest.containers {
		_, isFound, err := runtime.Inspect(ctx, c.name)
		if err != nil {
			return rerrors.Wrap(err, "error inspecting container")
		}

		if isFound {
			return rerrors.Wrap(errContainerExists, c.name)
		}
	}

	for _, it := range manifest.items {
		if it.kind != velez_api.TransferItemKind_TRANSFER_ITEM_KIND_VOLUME {
			continue
		}

		err := s.ensureVolumeAbsent(ctx, it.name)
		if err != nil {
			return rerrors.Wrap(err)
		}
	}

	return nil
}

func (s *TransferInServiceImpl) ensureVolumeAbsent(ctx context.Context, name string) error {
	_, err := s.docker.Client().VolumeInspect(ctx, name)
	if err == nil {
		return rerrors.Wrap(errVolumeExists, name)
	}

	if errdefs.IsNotFound(err) {
		return nil
	}

	return rerrors.Wrap(err, "error inspecting volume")
}
