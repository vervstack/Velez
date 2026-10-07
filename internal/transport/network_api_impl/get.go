package network_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) GetNetwork(
	ctx context.Context,
	req *pb.GetNetwork_Request,
) (*pb.GetNetwork_Response, error) {
	resp, err := impl.networks.GetNetwork(ctx, req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting network")
	}

	return resp, nil
}
