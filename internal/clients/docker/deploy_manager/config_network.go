package deploy_manager

import (
	"github.com/docker/docker/api/types/network"

	"github.com/godverv/Velez/internal/backservice/env"
	"github.com/godverv/Velez/pkg/velez_api"
)

func getNetworkConfig(req *velez_api.CreateSmerd_Request) (networkConfig *network.NetworkingConfig) {
	networkConfig = &network.NetworkingConfig{}

	networkConfig.EndpointsConfig = make(map[string]*network.EndpointSettings)

	vervNetwork := &network.EndpointSettings{
		Aliases: []string{req.GetName()},
	}
	networkConfig.EndpointsConfig[env.VervNetwork] = vervNetwork

	return networkConfig
}
