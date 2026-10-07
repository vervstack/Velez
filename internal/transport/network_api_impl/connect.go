package network_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) ConnectContainer(
	ctx context.Context,
	req *pb.ConnectContainer_Request,
) (*pb.ConnectContainer_Response, error) {
	resp, err := impl.networks.ConnectContainer(ctx, req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error connecting container to network")
	}

	return resp, nil
}
