package transfer_api_impl

import (
	"context"

	"github.com/google/uuid"
	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) CancelTransfer(
	ctx context.Context,
	req *pb.CancelTransfer_Request,
) (*pb.CancelTransfer_Response, error) {
	id, err := uuid.Parse(req.GetTransferId())
	if err != nil {
		return nil, rerrors.Wrap(errInvalidTransferId)
	}

	err = impl.transfersOut.Cancel(ctx, id)
	if err != nil {
		return nil, rerrors.Wrap(err, "error cancelling transfer")
	}

	return &pb.CancelTransfer_Response{}, nil
}
