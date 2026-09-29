package parser

import (
	"strconv"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"go.redsock.ru/toolbox"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func FromPorts(settings *velez_api.Container_Settings) map[nat.Port][]nat.PortBinding {
	if settings == nil {
		return nil
	}

	out := make(map[nat.Port][]nat.PortBinding, len(settings.GetPorts()))

	for _, item := range settings.GetPorts() {
		if item.ExposedTo == nil {
			// TODO auto asigne if not exists
			continue
		}

		if item.GetProtocol() == velez_api.Port_unknown {
			item.Protocol = velez_api.Port_tcp
		}

		servicePortStr := strconv.FormatUint(uint64(item.GetServicePortNumber()), 10)
		containerPort, _ := nat.NewPort(item.GetProtocol().String(), servicePortStr)

		out[containerPort] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: strconv.FormatUint(uint64(item.GetExposedTo()), 10),
			},
		}
	}

	return out
}

// ToPortsMapping reads the configured HostConfig.PortBindings. A binding with
// an empty HostPort (`docker run -p 80`) asks the daemon for a random host
// port, so its real one is taken from the running container's actual
// NetworkSettings.Ports; when that is unknown too (container not running) the
// port is left unassigned.
func ToPortsMapping(configured, actual map[nat.Port][]nat.PortBinding) []*velez_api.Port {
	if len(configured) == 0 {
		return nil
	}

	out := make([]*velez_api.Port, 0, len(configured))

	for contPort, hostPorts := range configured {
		for idx, hostPort := range hostPorts {
			binding := &velez_api.Port{
				ServicePortNumber: uint32(contPort.Int()),
				Protocol:          velez_api.Port_Protocol(velez_api.Port_Protocol_value[contPort.Proto()]),
			}

			exposedTo := resolveHostPort(hostPort, actual[contPort], idx)
			if exposedTo != 0 {
				binding.ExposedTo = &exposedTo
			}

			out = append(out, binding)
		}
	}

	return out
}

func resolveHostPort(configured nat.PortBinding, actual []nat.PortBinding, idx int) uint32 {
	hostPort := configured.HostPort

	if hostPort == "" && idx < len(actual) {
		hostPort = actual[idx].HostPort
	}

	port, err := strconv.ParseUint(hostPort, 10, 32)
	if err != nil {
		return 0
	}

	return uint32(port)
}

func ToPortsSlice(ports []container.Port) []*velez_api.Port {
	out := make([]*velez_api.Port, 0, len(ports))

	uniquePublicPort := map[uint32]struct{}{}

	for _, p := range ports {
		newP := ToPort(p)

		if newP.ExposedTo != nil {
			_, alreadyExists := uniquePublicPort[newP.GetExposedTo()]
			if alreadyExists {
				continue
			}

			uniquePublicPort[newP.GetExposedTo()] = struct{}{}
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
