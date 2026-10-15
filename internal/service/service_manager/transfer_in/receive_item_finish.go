package transfer_in

import (
	"go.redsock.ru/rerrors"
)

// finishCurrent applies the open item, if any, and reports it to the peer.
func (r *itemReceiver) finishCurrent() error {
	if r.current == nil {
		return nil
	}

	applier := r.current

	r.current = nil

	err := applier.finish()
	if err != nil {
		return rerrors.Wrap(err)
	}

	err = r.stream.Send(newItemAppliedResponse(applier.item.name))
	if err != nil {
		return rerrors.Wrap(err, "error sending item applied")
	}

	return nil
}

func (r *itemReceiver) end() error {
	err := r.finishCurrent()
	if err != nil {
		return rerrors.Wrap(err)
	}

	if len(r.seen) != len(r.run.manifest.items) {
		return rerrors.Wrap(errItemsIncomplete)
	}

	return nil
}

func (r *itemReceiver) abortCurrent(cause error) {
	if r.current == nil {
		return
	}

	r.current.abort(cause)

	r.current = nil
}
