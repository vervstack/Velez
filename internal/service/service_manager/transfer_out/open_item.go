package transfer_out

import (
	"context"
	"io"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (s *Service) openItem(ctx context.Context, item *velez_api.Transfer_ItemManifest) (io.ReadCloser, error) {
	switch item.GetKind() {
	case velez_api.TransferItemKind_TRANSFER_ITEM_KIND_IMAGE:
		reader, err := s.transferIO.ImageSave(ctx, item.GetName())
		if err != nil {
			return nil, rerrors.Wrap(err, "error saving image")
		}

		return reader, nil
	case velez_api.TransferItemKind_TRANSFER_ITEM_KIND_VOLUME:
		reader, err := s.transferIO.VolumeRead(ctx, item.GetName())
		if err != nil {
			return nil, rerrors.Wrap(err, "error reading volume")
		}

		return reader, nil
	case velez_api.TransferItemKind_TRANSFER_ITEM_KIND_UNSPECIFIED:
		return nil, rerrors.Wrap(errUnknownItemKind)
	default:
		return nil, rerrors.Wrap(errUnknownItemKind)
	}
}
