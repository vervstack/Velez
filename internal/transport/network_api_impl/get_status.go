package network_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) GetNetworkStatus(
	ctx context.Context,
	req *pb.GetNetworkStatus_Request,
) (*pb.GetNetworkStatus_Response, error) {
	resp, err := impl.networks.GetStatus(ctx, req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting network status")
	}

	return resp, nil
}
