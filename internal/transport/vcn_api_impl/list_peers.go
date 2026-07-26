package vcn_api_impl

import (
	"context"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) ListPeers(_ context.Context, _ *pb.ListPeers_Request) (*pb.ListPeers_Response, error) {
	// Return empty list for now; headscale integration is future work
	return &pb.ListPeers_Response{
		Peers: []*pb.VcnPeer{},
	}, nil
}
