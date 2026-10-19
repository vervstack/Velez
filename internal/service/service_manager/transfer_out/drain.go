package transfer_out

import (
	"errors"
	"io"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

// responseDrain reads the target's frames concurrently with the sender, so
// flow control on the stream can never deadlock the two directions.
type responseDrain struct {
	done chan struct{}
	err  error
}

func startDrain(stream velez_api.TransferAPI_TransferClient) *responseDrain {
	drain := &responseDrain{done: make(chan struct{})}

	go drain.run(stream)

	return drain
}

func (d *responseDrain) run(stream velez_api.TransferAPI_TransferClient) {
	defer close(d.done)

	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			d.err = rerrors.Wrap(errTargetClosedEarly)

			return
		}

		if err != nil {
			d.err = peerError(err, "error receiving from target")

			return
		}

		isFinished, frameErr := checkFrame(resp)
		if frameErr != nil {
			d.err = rerrors.Wrap(frameErr)

			return
		}

		if isFinished {
			return
		}
	}
}

// wait blocks until the target finished or the stream broke; nil means Finished arrived.
func (d *responseDrain) wait() error {
	<-d.done

	if d.err != nil {
		return rerrors.Wrap(d.err)
	}

	return nil
}

func checkFrame(resp *velez_api.Transfer_Response) (bool, error) {
	if resp.GetFinished() != nil {
		return true, nil
	}

	if resp.GetItemApplied() != nil {
		return false, nil
	}

	return false, rerrors.Wrap(errUnexpectedFrame)
}
