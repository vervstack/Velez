package transfer_in

import (
	"context"
	"io"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/transfer_io"
)

// itemApplier feeds one item's batches through a pipe into the loader that
// applies it, so an item never has to fit in memory.
type itemApplier struct {
	item       manifestItem
	transferIO transfer_io.TransferIO
	reader     *io.PipeReader
	writer     *io.PipeWriter
	done       chan error
}

func startItemApplier(ctx context.Context, item manifestItem, transferIO transfer_io.TransferIO) *itemApplier {
	reader, writer := io.Pipe()

	applier := &itemApplier{
		item:       item,
		transferIO: transferIO,
		reader:     reader,
		writer:     writer,
		done:       make(chan error, 1),
	}

	go applier.run(ctx)

	return applier
}

func (a *itemApplier) run(ctx context.Context) {
	err := a.apply(ctx)

	// A failed loader unblocks a writer stuck in Write with the loader's error.
	_ = a.reader.CloseWithError(err)

	a.done <- err
}

func (a *itemApplier) apply(ctx context.Context) error {
	var err error

	switch a.item.kind {
	case velez_api.TransferItemKind_TRANSFER_ITEM_KIND_IMAGE:
		err = a.transferIO.ImageLoad(ctx, a.reader)
	case velez_api.TransferItemKind_TRANSFER_ITEM_KIND_VOLUME:
		err = a.transferIO.VolumeWrite(ctx, a.item.name, a.reader)
	case velez_api.TransferItemKind_TRANSFER_ITEM_KIND_UNSPECIFIED:
		err = rerrors.Wrap(errInvalidItem, a.item.name)
	}

	if err != nil {
		return rerrors.Wrap(err)
	}

	return nil
}

func (a *itemApplier) write(data []byte) error {
	_, err := a.writer.Write(data)
	if err != nil {
		return rerrors.Wrap(err, "error writing item batch")
	}

	return nil
}

// finish ends the item's stream and waits until the loader has applied it.
func (a *itemApplier) finish() error {
	_ = a.writer.Close()

	err := <-a.done
	if err != nil {
		return rerrors.Wrap(err, "error applying item")
	}

	return nil
}

func (a *itemApplier) abort(cause error) {
	_ = a.writer.CloseWithError(cause)

	<-a.done
}
