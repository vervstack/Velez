package parser

import (
	"strconv"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"go.redsock.ru/toolbox"

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

func ToPortsMapping(ports map[nat.Port][]nat.PortBinding) []*velez_api.Port {
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

func ToPortsSlice(ports []container.Port) []*velez_api.Port {
	out := make([]*velez_api.Port, 0, len(ports))

	uniquePublicPort := map[uint32]struct{}{}

	for _, p := range ports {

		newP := ToPort(p)

		if newP.ExposedTo != nil {
			_, alreadyExists := uniquePublicPort[*newP.ExposedTo]
			if alreadyExists {
				continue
			}

			uniquePublicPort[*newP.ExposedTo] = struct{}{}
		}

		out = append(out, newP)
	}

	return out
}

func ToPort(port container.Port) *velez_api.Port {
	return &velez_api.Port{
		ServicePortNumber: uint32(port.PrivatePort),
		Protocol:          ToPortProtocol(port.Type),
		ExposedTo:         toolbox.ToPtr(uint32(port.PublicPort)),
	}
}

func ToPortProtocol(tp string) velez_api.Port_Protocol {
	switch tp {
	case "tcp":
		return velez_api.Port_tcp
	case "udp":
		return velez_api.Port_udp
	default:
		return velez_api.Port_unknown
	}
}
