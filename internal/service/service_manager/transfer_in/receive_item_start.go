package transfer_in

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (r *itemReceiver) startItem(ctx context.Context, name string) error {
	err := r.finishCurrent()
	if err != nil {
		return rerrors.Wrap(err)
	}

	item, isFound := r.run.manifest.item(name)
	if !isFound {
		return rerrors.Wrap(errUnknownItem, name)
	}

	_, isSeen := r.seen[name]
	if isSeen {
		return rerrors.Wrap(errItemRepeated, name)
	}

	r.seen[name] = struct{}{}

	if item.kind == velez_api.TransferItemKind_TRANSFER_ITEM_KIND_VOLUME {
		r.run.createdVolumes = append(r.run.createdVolumes, name)
	}

	r.current = startItemApplier(ctx, item, r.service.transferIO)

	return nil
}
