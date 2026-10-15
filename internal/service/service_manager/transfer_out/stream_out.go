package transfer_out

import (
	"context"
	"errors"
	"io"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/utils/common"
)

// outbound is the sending half of a session; only the transfer goroutine writes to it.
type outbound struct {
	stream  velez_api.TransferAPI_TransferClient
	drain   *responseDrain
	counter *progress
}

// streamItems sends every item, then End, and waits for the target's Finished.
func (s *Service) streamItems(
	ctx context.Context,
	sess *session,
	items []*velez_api.Transfer_ItemManifest,
	counter *progress,
) error {
	out := &outbound{
		stream:  sess.stream,
		drain:   startDrain(sess.stream),
		counter: counter,
	}

	for _, item := range items {
		err := s.sendItem(ctx, out, item)
		if err != nil {
			return rerrors.Wrap(err)
		}
	}

	err := out.send(endRequest())
	if err != nil {
		return rerrors.Wrap(err)
	}

	err = sess.stream.CloseSend()
	if err != nil {
		return rerrors.Wrap(err, "error closing send side")
	}

	err = out.drain.wait()
	if err != nil {
		return rerrors.Wrap(err)
	}

	return nil
}

func (s *Service) sendItem(ctx context.Context, out *outbound, item *velez_api.Transfer_ItemManifest) error {
	reader, err := s.openItem(ctx, item)
	if err != nil {
		return rerrors.Wrap(err)
	}
	defer common.CloseWithLog(reader.Close, "transfer item reader")

	err = out.send(itemStartRequest(item.GetName()))
	if err != nil {
		return rerrors.Wrap(err)
	}

	err = forEachChunk(reader, chunkSize, out.sendChunk)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return nil
}

func (o *outbound) sendChunk(chunk []byte) error {
	err := o.send(batchRequest(chunk))
	if err != nil {
		return rerrors.Wrap(err)
	}

	o.counter.doneBytes.Add(int64(len(chunk)))

	return nil
}

// send reports the target's own verdict when the stream ended under the sender: Send itself only says io.EOF.
func (o *outbound) send(req *velez_api.Transfer_Request) error {
	err := o.stream.Send(req)
	if err == nil {
		return nil
	}

	if !errors.Is(err, io.EOF) {
		return peerError(err, "error sending frame")
	}

	cause := o.drain.wait()
	if cause == nil {
		cause = errTargetClosedEarly
	}

	return rerrors.Wrap(cause)
}
