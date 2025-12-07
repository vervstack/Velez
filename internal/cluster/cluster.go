package cluster

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	headscaleClient "go.vervstack.ru/Velez/internal/clients/cluster_clients/headscale"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients/makosh"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients/matreshka"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/cluster/configuration"
	"go.vervstack.ru/Velez/internal/cluster/env"
	"go.vervstack.ru/Velez/internal/cluster/service_discovery"
	"go.vervstack.ru/Velez/internal/cluster/verv_private_network"
	"go.vervstack.ru/Velez/internal/config"
)

type clusterClients struct {
	matreshka        matreshka.Client
	serviceDiscovery *makosh.ServiceDiscovery
	headscale        *headscaleClient.Client
}

func Setup(ctx context.Context, cfg config.Config, nodeClients node_clients.NodeClients) (cluster_clients.ClusterClients, error) {
	err := env.SetupEnvironment(nodeClients)
	if err != nil {
		return nil, rerrors.Wrap(err, "error settings up environment")
	}

	// TODO VPN must be presented in one of options
	// 1. Single node - docker network (❌not implemented)
	// 2. Multi node Tailscale - using 3rd party service (❌not implemented)
	// 3. Multi node - using headscale setup (⚠️ in development)
	vcnClient, err := verv_private_network.LaunchHeadscale(ctx, cfg, nodeClients)
	if err != nil {
		return nil, rerrors.Wrap(err, "error during vpn server client initialization")
	}

	sdClient, err := service_discovery.SetupMakosh(ctx, cfg, nodeClients, vcnClient)
	if err != nil {
		return nil, rerrors.Wrap(err, "error during makosh setup")
	}

	var cfgClient matreshka.Client
	if cfg.Environment.MatreshkaIsEnabled {
		cfgClient, err = configuration.SetupMatreshka(ctx, cfg, nodeClients, sdClient, vcnClient)
		if err != nil {
			return nil, rerrors.Wrap(err, "error during matreshka setup")
		}
	}

	return &clusterClients{
		cfgClient,
		sdClient,
		vcnClient,
	}, nil
}

func (c *clusterClients) Configurator() cluster_clients.Configurator {
	if c.matreshka == nil {
		return &disabledConfigurator{}
	}

	return c.matreshka
}

func (c *clusterClients) Vpn() cluster_clients.VervClosedNetworkClient {
	if c.headscale == nil {
		return &disabledVpn{}
	}
	return c.headscale
}

func (c *clusterClients) ServiceDiscovery() cluster_clients.ServiceDiscovery {
	if c.serviceDiscovery == nil {
		return &disabledServiceDiscovery{}
	}

	return c.serviceDiscovery
}
