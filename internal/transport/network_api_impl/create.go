package network_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) CreateNetwork(
	ctx context.Context,
	req *pb.CreateNetwork_Request,
) (*pb.CreateNetwork_Response, error) {
	resp, err := impl.networks.CreateNetwork(ctx, req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating network")
	}

	return resp, nil
}
