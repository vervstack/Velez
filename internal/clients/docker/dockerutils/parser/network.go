package parser

import (
	"github.com/docker/docker/api/types/network"

	"github.com/godverv/Velez/pkg/velez_api"
)

func FromNetworks(settings *velez_api.Container_Settings) map[string]*network.EndpointSettings {
	if settings == nil {
		return nil
	}

	res := make(map[string]*network.EndpointSettings)

	for _, n := range settings.GetNetworks() {
		res[n.GetNetworkName()] = &network.EndpointSettings{
			Aliases: n.GetAliases(),
		}
	}

	return res
}
