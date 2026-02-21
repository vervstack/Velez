package cluster

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients/makosh"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients/matreshka"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients/state"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/cluster/cluster_state"
	"go.vervstack.ru/Velez/internal/cluster/configuration"
	"go.vervstack.ru/Velez/internal/cluster/env"
	"go.vervstack.ru/Velez/internal/cluster/service_discovery"
	"go.vervstack.ru/Velez/internal/cluster/verv_closed_network"
	"go.vervstack.ru/Velez/internal/config"
)

type clusterClients struct {
	matreshka        matreshka.Client
	serviceDiscovery *makosh.ServiceDiscovery
	headscale        cluster_clients.VervClosedNetworkClient
	stateManager     cluster_clients.ClusterStateManagerContainer
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
	vcnClient, err := verv_closed_network.SetupVcn(ctx, nodeClients)
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

	clusterStateManagerContainer := state.NewContainer()

	localState := nodeClients.LocalStateManager().Get()
	if localState.ClusterState.PgRootDsn != "" {
		err = cluster_state.SetupMasterPg(ctx, nodeClients)
		if err != nil {
			return nil, rerrors.Wrap(err, "error setting up master postgres for cluster state")
		}
	}

	//
	//// TODO make separate state manager for root user
	//pg, err := NewPgStateManager(state.ClusterState.PgNodeDsn)
	//if err != nil {
	//	log.Err(err).
	//		Msg("error connecting to PgRootDsn")
	//} else {
	//	sm.Set(pg)
	//}

	return &clusterClients{
		cfgClient,
		sdClient,
		vcnClient,
		clusterStateManagerContainer,
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
		return &verv_closed_network.DisabledVcnImpl{}
	}
	return c.headscale
}

func (c *clusterClients) ServiceDiscovery() cluster_clients.ServiceDiscovery {
	if c.serviceDiscovery == nil {
		return &disabledServiceDiscovery{}
	}

	return c.serviceDiscovery
}

func (c *clusterClients) StateManager() cluster_clients.ClusterStateManagerContainer {
	return c.stateManager
}
