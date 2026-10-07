package network_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) DeleteNetwork(
	ctx context.Context,
	req *pb.DeleteNetwork_Request,
) (*pb.DeleteNetwork_Response, error) {
	resp, err := impl.networks.DeleteNetwork(ctx, req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error deleting network")
	}

	return resp, nil
}
