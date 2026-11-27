package vcn_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service/user_errors"
)

func (impl *Impl) ListPeers(ctx context.Context, req *pb.ListPeers_Request) (*pb.ListPeers_Response, error) {
	var (
		nodes []domain.VcnNode
		err   error
	)

	switch {
	case req.GetNetworkId() != "":
		nodes, err = impl.networks.Peers(ctx, req.GetNetworkId())
	case impl.vpnService != nil:
		nodes, err = impl.vpnService.ListNodes(ctx)
		if rerrors.Is(err, user_errors.ErrServiceIsDisabled) {
			err = nil
		}
	}

	if err != nil {
		return nil, rerrors.Wrap(err, "error listing VPN peers")
	}

	return &pb.ListPeers_Response{Peers: toPeers(nodes)}, nil
}
