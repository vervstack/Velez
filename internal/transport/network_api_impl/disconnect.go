package network_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) DisconnectContainer(
	ctx context.Context,
	req *pb.DisconnectContainer_Request,
) (*pb.DisconnectContainer_Response, error) {
	resp, err := impl.networks.DisconnectContainer(ctx, req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error disconnecting container from network")
	}

	return resp, nil
}
