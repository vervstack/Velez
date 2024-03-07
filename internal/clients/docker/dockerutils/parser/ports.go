package parser

import (
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"

	"github.com/godverv/Velez/pkg/velez_api"
)

func FromPorts(settings *velez_api.Container_Settings) map[nat.Port][]nat.PortBinding {
	if settings == nil {
		return nil
	}

	out := make(map[nat.Port][]nat.PortBinding, len(settings.Ports))

	for _, item := range settings.Ports {
		if item.Protoc == velez_api.PortBindings_unknown {
			item.Protoc = velez_api.PortBindings_tcp
		}

		containerPort, _ := nat.NewPort(item.Protoc.String(), strconv.FormatUint(uint64(item.Container), 10))

		out[containerPort] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: strconv.FormatUint(uint64(item.Host), 10),
			},
		}
	}

	return out
}

func ToPorts(ports []types.Port) []*velez_api.PortBindings {
	out := make([]*velez_api.PortBindings, len(ports))

	for i, item := range ports {
		out[i] = &velez_api.PortBindings{
			Host:      uint32(item.PublicPort),
			Container: uint32(item.PrivatePort),
			Protoc:    velez_api.PortBindings_Protocol(velez_api.PortBindings_Protocol_value[item.Type]),
		}
	}
	return out
}
