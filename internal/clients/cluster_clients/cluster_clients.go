package cluster_clients

import (
	"context"

	"go.redsock.ru/rerrors"
	"go.vervstack.ru/matreshka/pkg/app/matreshka_client"
	"google.golang.org/grpc"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients/headscale"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients/makosh"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients/matreshka"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/config"
)

type ClusterClients interface {
	Configurator() Configurator
	Vpn() VervPrivateNetworkClient
}

type clusterClients struct {
	matreshka        matreshka.Client
	serviceDiscovery makosh.ServiceDiscovery
	headscale        *headscale.Client
}

func NewClusterClients(ctx context.Context, cfg config.Config, nodeClients node_clients.NodeClients) (ClusterClients, error) {
	serviceDiscovery, err := makosh.NewServiceDiscovery(cfg)
	if err != nil {
		return nil, rerrors.Wrap(err, "error initializing service discovery ")
	}

	matreshkaClient, err := matreshka.NewClient(
		grpc.WithUnaryInterceptor(
			matreshka_client.WithHeader(
				matreshka_client.Pass, nodeClients.LocalStateManager().Get().MatreshkaKey)))
	if err != nil {
		return nil, rerrors.Wrap(err, "error initializing matreshka client ")
	}

	headscaleClient, err := headscale.New(ctx, nodeClients, "headscale")
	if err != nil {
		return nil, rerrors.Wrap(err, "error during vpn server client initialization")
	}

	return &clusterClients{
		matreshkaClient,
		serviceDiscovery,
		headscaleClient,
	}, nil
}

func (c *clusterClients) Configurator() Configurator {
	return c.matreshka
}

func (c *clusterClients) Vpn() VervPrivateNetworkClient {
	return c.headscale
}
