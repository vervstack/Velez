package deploy_manager

import (
	"github.com/docker/docker/api/types/network"

	"github.com/godverv/Velez/internal/backservice/env"
	"github.com/godverv/Velez/pkg/velez_api"
)

func getNetworkConfig(req *velez_api.CreateSmerd_Request) (networkConfig *network.NetworkingConfig) {
	networkConfig = &network.NetworkingConfig{}

	if len(req.Settings.Ports) == 0 {
		return networkConfig
	}

	networkConfig.EndpointsConfig = make(map[string]*network.EndpointSettings)

	vervNetwork := &network.EndpointSettings{
		Aliases: []string{req.GetName()},
	}
	networkConfig.EndpointsConfig[env.VervNetwork] = vervNetwork

	// required in order to expose ports on some platforms (e.g. orbs)
	networkConfig.EndpointsConfig["bridge"] = &network.EndpointSettings{}

	return networkConfig
}
