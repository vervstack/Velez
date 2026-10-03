package jobs

import (
	"slices"

	"github.com/docker/docker/api/types/container"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils/parser"
)

// findForeignOccupiedPort returns the first requested host port that is bound by something other than the
// container itself.
func findForeignOccupiedPort(requested []*velez_api.Port, occupied []uint32, own []*velez_api.Port) (uint32, bool) {
	for _, p := range requested {
		if p.ExposedTo == nil {
			continue
		}

		port := p.GetExposedTo()

		isOccupied := slices.Contains(occupied, port)
		if !isOccupied {
			continue
		}

		isOwn := slices.ContainsFunc(own, func(o *velez_api.Port) bool {
			return o.ExposedTo != nil && o.GetExposedTo() == port
		})
		if isOwn {
			continue
		}

		return port, true
	}

	return 0, false
}

// currentHostPorts lists the host ports the container publishes right now.
func currentHostPorts(info container.InspectResponse) []*velez_api.Port {
	if info.HostConfig == nil {
		return nil
	}

	return parser.ToPortsFromInspect(info)
}
