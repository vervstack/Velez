package network_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) ListNetworks(
	ctx context.Context,
	req *pb.ListNetworks_Request,
) (*pb.ListNetworks_Response, error) {
	resp, err := impl.networks.ListNetworks(ctx, req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing networks")
	}

	return resp, nil
}
