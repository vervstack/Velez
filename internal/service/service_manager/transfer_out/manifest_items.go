package transfer_out

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

// itemManifests lists the images, then the volumes, the transfer carries, with their sizes.
func (s *TransferOutServiceImpl) itemManifests(
	ctx context.Context,
	containers []container.InspectResponse,
) ([]*velez_api.Transfer_ItemManifest, error) {
	images := distinctImages(containers)
	volumes := distinctVolumes(containers)
	items := make([]*velez_api.Transfer_ItemManifest, 0, len(images)+len(volumes))

	for _, ref := range images {
		inspected, err := s.docker.Client().ImageInspect(ctx, ref)
		if err != nil {
			return nil, rerrors.Wrap(err, "error inspecting image")
		}

		items = append(items, newItemManifest(ref, velez_api.TransferItemKind_TRANSFER_ITEM_KIND_IMAGE, inspected.Size))
	}

	for _, name := range volumes {
		size, err := s.transferIO.VolumeSize(ctx, name)
		if err != nil {
			return nil, rerrors.Wrap(err, "error measuring volume")
		}

		items = append(items, newItemManifest(name, velez_api.TransferItemKind_TRANSFER_ITEM_KIND_VOLUME, size))
	}

	return items, nil
}

func newItemManifest(name string, kind velez_api.TransferItemKind, totalBytes int64) *velez_api.Transfer_ItemManifest {
	return &velez_api.Transfer_ItemManifest{
		Name:       name,
		Kind:       kind,
		TotalBytes: totalBytes,
	}
}

func totalBytes(items []*velez_api.Transfer_ItemManifest) int64 {
	var total int64

	for _, item := range items {
		total += item.GetTotalBytes()
	}

	return total
}
