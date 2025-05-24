package parser

import (
	"strconv"

	"github.com/docker/go-connections/nat"

	"go.vervstack.ru/Velez/pkg/velez_api"
)

func FromPorts(settings *velez_api.Container_Settings) map[nat.Port][]nat.PortBinding {
	if settings == nil {
		return nil
	}

	out := make(map[nat.Port][]nat.PortBinding, len(settings.Ports))

	for _, item := range settings.Ports {
		if item.ExposedTo == nil {
			// TODO auto asigne if not exists
			continue
		}
		if item.Protocol == velez_api.Port_unknown {
			item.Protocol = velez_api.Port_tcp
		}

		servicePortStr := strconv.FormatUint(uint64(item.ServicePortNumber), 10)
		containerPort, _ := nat.NewPort(item.Protocol.String(), servicePortStr)

		out[containerPort] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: strconv.FormatUint(uint64(*item.ExposedTo), 10),
			},
		}
	}

	return out
}

func ToPorts(ports map[nat.Port][]nat.PortBinding) []*velez_api.Port {
	if len(ports) == 0 {
		return nil
	}

	out := make([]*velez_api.Port, 0, len(ports))

	for contPort, hostPorts := range ports {
		for _, hostPort := range hostPorts {
			port, _ := strconv.ParseUint(hostPort.HostPort, 10, 64)
			port32 := uint32(port)

			binding := &velez_api.Port{
				ExposedTo:         &port32,
				ServicePortNumber: uint32(contPort.Int()),
				Protocol:          velez_api.Port_Protocol(velez_api.Port_Protocol_value[contPort.Proto()]),
			}

			out = append(out, binding)
		}

	}

	return out
}
