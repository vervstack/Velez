package transfer_in

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

// readBegin is the authorization gate: the transfer stream bypasses the
// Authorization interceptor, so the single-use token consumed here is the only
// thing that lets a peer in.
func (s *TransferInServiceImpl) readBegin(
	ctx context.Context,
	stream velez_api.TransferAPI_TransferServer,
) (*velez_api.Transfer_Begin, error) {
	req, err := stream.Recv()
	if err != nil {
		return nil, rerrors.Wrap(err, "error receiving begin frame")
	}

	begin := req.GetBegin()
	if begin == nil {
		return nil, rerrors.Wrap(errBeginExpected)
	}

	isConsumed, err := s.dataStorage.Transfers().ConsumeToken(ctx, hashToken(begin.GetToken()))
	if err != nil {
		return nil, rerrors.Wrap(err, "error consuming token")
	}

	if !isConsumed {
		return nil, rerrors.Wrap(errTokenInvalid)
	}

	return begin, nil
}
