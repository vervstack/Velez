package parser

import (
	"strconv"

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

func ToPorts(ports map[nat.Port][]nat.PortBinding) []*velez_api.PortBindings {
	if len(ports) == 0 {
		return nil
	}

	out := make([]*velez_api.PortBindings, 0, len(ports))

	for contPort, hostPorts := range ports {
		for _, hostPort := range hostPorts {
			port, _ := strconv.ParseUint(hostPort.HostPort, 10, 64)
			binding := &velez_api.PortBindings{
				Host:      uint32(port),
				Container: uint32(contPort.Int()),
				Protoc:    velez_api.PortBindings_Protocol(velez_api.PortBindings_Protocol_value[contPort.Proto()]),
			}

			out = append(out, binding)
		}

	}

	return out
}
