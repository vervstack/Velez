package vcn_api_impl

import (
	"context"
	"strings"
	"time"

	"go.redsock.ru/rerrors"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/transport/common"
)

func (impl *Impl) CreateNetwork(ctx context.Context, req *pb.CreateVcnNetwork_Request) (
	*pb.CreateVcnNetwork_Response, error,
) {
	domainReq := common.CreateVcnNetworkRequest(ctx, req)

	network, err := impl.networks.Create(ctx, domainReq)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	return &pb.CreateVcnNetwork_Response{Network: toNetwork(network)}, nil
}

func (impl *Impl) ListNetworks(
	ctx context.Context, _ *pb.ListVcnNetworks_Request,
) (*pb.ListVcnNetworks_Response, error) {
	networks, err := impl.networks.List(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	response := &pb.ListVcnNetworks_Response{Networks: make([]*pb.VcnNetwork, 0, len(networks))}

	for _, network := range networks {
		response.Networks = append(response.Networks, toNetwork(network))
	}

	return response, nil
}

func (impl *Impl) GetNetwork(ctx context.Context, req *pb.GetVcnNetwork_Request) (*pb.GetVcnNetwork_Response, error) {
	details, err := impl.networks.Get(ctx, req.GetId())
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	response := &pb.GetVcnNetwork_Response{
		Network: toNetwork(details.Network),
		Peers:   toPeers(details.Peers),
		ApiKeys: make([]*pb.NetworkApiKey, 0, len(details.ApiKeys)),
	}

	for _, key := range details.ApiKeys {
		response.ApiKeys = append(response.ApiKeys, toApiKey(key))
	}

	return response, nil
}

func (impl *Impl) DropNetwork(
	ctx context.Context, req *pb.DropVcnNetwork_Request,
) (*pb.DropVcnNetwork_Response, error) {
	err := impl.networks.Drop(ctx, req.GetId())
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	return &pb.DropVcnNetwork_Response{}, nil
}

func (impl *Impl) IssueNetworkApiKey(ctx context.Context, req *pb.IssueNetworkApiKey_Request) (
	*pb.IssueNetworkApiKey_Response, error,
) {
	key, err := impl.networks.IssueApiKey(ctx, req.GetNetworkId())
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	prefix, _, _ := strings.Cut(key, ".")
	metadata := &pb.NetworkApiKey{Prefix: prefix}

	return &pb.IssueNetworkApiKey_Response{Key: key, ApiKey: metadata}, nil
}

func (impl *Impl) ExpireNetworkApiKey(ctx context.Context, req *pb.ExpireNetworkApiKey_Request) (
	*pb.ExpireNetworkApiKey_Response, error,
) {
	err := impl.networks.ExpireApiKey(ctx, req.GetNetworkId(), req.GetPrefix())
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	return &pb.ExpireNetworkApiKey_Response{}, nil
}

func toNetwork(record domain.Network) *pb.VcnNetwork {
	return &pb.VcnNetwork{
		Id:          record.Id,
		Name:        record.Name,
		Provider:    record.Provider,
		Source:      record.Source,
		Url:         record.Url,
		ServiceName: record.ServiceName,
		CreatedAt:   toTimestamp(record.CreatedAt),
	}
}

func toApiKey(key domain.VcnApiKey) *pb.NetworkApiKey {
	return &pb.NetworkApiKey{
		Prefix:     key.Prefix,
		CreatedAt:  toTimestamp(key.CreatedAt),
		Expiration: toTimestamp(key.Expiration),
		LastSeen:   toTimestamp(key.LastSeen),
	}
}

func toTimestamp(value time.Time) *timestamppb.Timestamp {
	if value.IsZero() {
		return nil
	}

	return timestamppb.New(value)
}

func toPeers(nodes []domain.VcnNode) []*pb.VcnPeer {
	peers := make([]*pb.VcnPeer, 0, len(nodes))

	for _, node := range nodes {
		status := "offline"
		if node.IsOnline {
			status = "online"
		}

		peer := &pb.VcnPeer{
			Id:          node.Id,
			Host:        node.Name,
			Type:        "client",
			Status:      status,
			IpAddresses: node.IpAddresses,
		}

		peers = append(peers, peer)
	}

	return peers
}
