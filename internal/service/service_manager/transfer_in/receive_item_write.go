package transfer_in

import (
	"context"

	"go.redsock.ru/rerrors"
)

func (r *itemReceiver) writeBatch(ctx context.Context, data []byte) error {
	if r.current == nil {
		return rerrors.Wrap(errNoOpenItem)
	}

	err := r.current.write(data)
	if err != nil {
		return rerrors.Wrap(err)
	}

	r.progress.add(ctx, len(data))

	return nil
}
