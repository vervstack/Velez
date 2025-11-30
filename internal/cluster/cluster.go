package cluster

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients/headscale"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients/makosh"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients/matreshka"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/cluster/configuration"
	"go.vervstack.ru/Velez/internal/cluster/env"
	"go.vervstack.ru/Velez/internal/cluster/service_discovery"
	"go.vervstack.ru/Velez/internal/config"
)

type Cluster interface {
	Configurator() cluster_clients.Configurator
	Vpn() cluster_clients.VervPrivateNetworkClient
	ServiceDiscovery() cluster_clients.ServiceDiscovery
}

type clusterClients struct {
	matreshka        matreshka.Client
	serviceDiscovery *makosh.ServiceDiscovery
	headscale        *headscale.Client
}

func Setup(ctx context.Context, cfg config.Config, nodeClients node_clients.NodeClients) (Cluster, error) {
	err := env.SetupEnvironment(nodeClients)
	if err != nil {
		return nil, rerrors.Wrap(err, "error settings up environment")
	}

	var headscaleClient *headscale.Client

	if cfg.Environment.VpnIsEnabled {
		headscaleClient, err = headscale.New(ctx, nodeClients, "headscale")
		if err != nil {
			return nil, rerrors.Wrap(err, "error during vpn server client initialization")
		}
	}

	var sdClient *makosh.ServiceDiscovery
	if cfg.Environment.MakoshIsEnabled && headscaleClient != nil {
		sdClient, err = service_discovery.SetupMakosh(ctx, cfg, nodeClients, headscaleClient)
		if err != nil {
			return nil, rerrors.Wrap(err, "error during makosh setup")
		}
	}

	var cfgClient matreshka.Client
	if cfg.Environment.MatreshkaIsEnabled {
		cfgClient, err = configuration.SetupMatreshka(ctx, cfg, nodeClients)
		if err != nil {
			return nil, rerrors.Wrap(err, "error during matreshka setup")
		}
	}

	return &clusterClients{
		cfgClient,
		sdClient,
		headscaleClient,
	}, nil
}

func (c *clusterClients) Configurator() cluster_clients.Configurator {
	if c.matreshka == nil {
		return &disabledConfigurator{}
	}

	return c.matreshka
}

func (c *clusterClients) Vpn() cluster_clients.VervPrivateNetworkClient {
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
