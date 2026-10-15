package transfer_api_impl

import (
	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) Transfer(stream pb.TransferAPI_TransferServer) error {
	err := impl.transfersIn.Receive(stream.Context(), stream)
	if err != nil {
		return rerrors.Wrap(err, "error receiving transfer")
	}

	return nil
}
