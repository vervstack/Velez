package transfer_in

import (
	"context"
	"errors"
	"io"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

type itemReceiver struct {
	service  *Service
	stream   velez_api.TransferAPI_TransferServer
	run      *receiveRun
	progress *progressTracker
	seen     map[string]struct{}
	current  *itemApplier
}

func (s *Service) receiveItems(
	ctx context.Context,
	stream velez_api.TransferAPI_TransferServer,
	run *receiveRun,
	progress *progressTracker,
) error {
	receiver := &itemReceiver{
		service:  s,
		stream:   stream,
		run:      run,
		progress: progress,
		seen:     make(map[string]struct{}, len(run.manifest.items)),
	}

	err := receiver.receiveAll(ctx)
	if err != nil {
		receiver.abortCurrent(err)

		return rerrors.Wrap(err)
	}

	return nil
}

func (r *itemReceiver) receiveAll(ctx context.Context) error {
	for {
		req, err := r.stream.Recv()
		if errors.Is(err, io.EOF) {
			return rerrors.Wrap(errStreamClosedEarly)
		}

		if err != nil {
			return rerrors.Wrap(err, "error receiving frame")
		}

		isEnd, err := r.handle(ctx, req)
		if err != nil {
			return rerrors.Wrap(err)
		}

		if isEnd {
			return nil
		}
	}
}

func (r *itemReceiver) handle(ctx context.Context, req *velez_api.Transfer_Request) (isEnd bool, err error) {
	switch frame := req.GetFrame().(type) {
	case *velez_api.Transfer_Request_ItemStart:
		err = r.startItem(ctx, frame.ItemStart.GetName())
	case *velez_api.Transfer_Request_Batch:
		err = r.writeBatch(ctx, frame.Batch.GetData())
	case *velez_api.Transfer_Request_End:
		return true, r.end()
	default:
		err = errUnexpectedFrame
	}

	if err != nil {
		return false, rerrors.Wrap(err)
	}

	return false, nil
}
