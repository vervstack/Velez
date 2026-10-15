package transfer_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

func (impl *Impl) StartTransfer(
	ctx context.Context,
	req *pb.StartTransfer_Request,
) (*pb.StartTransfer_Response, error) {
	startReq := domain.StartTransferReq{
		ContainerName:           req.GetContainerName(),
		SidecarNames:            req.GetSidecarNames(),
		Address:                 req.GetAddress(),
		Token:                   req.GetToken(),
		IsConsumersAcknowledged: req.GetIsConsumersAcknowledged(),
	}

	id, err := impl.transfersOut.Start(ctx, startReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error starting transfer")
	}

	resp := &pb.StartTransfer_Response{TransferId: id.String()}

	return resp, nil
}
