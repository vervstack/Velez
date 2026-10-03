package jobs

import (
	"github.com/docker/docker/api/types/container"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

// isStopOldFirstRequired reports whether the old container must be stopped before the new one starts:
// the request keeps its host ports, the new container reuses one of its host ports, or it has mounts
// the new container will share (two live processes on one data directory risk corruption).
func isStopOldFirstRequired(info container.InspectResponse, ports []*velez_api.Port, isKeepingPorts bool) bool {
	if isKeepingPorts {
		return true
	}

	if len(info.Mounts) > 0 {
		return true
	}

	current := currentHostPorts(info)

	for _, p := range ports {
		if p.ExposedTo == nil {
			continue
		}

		for _, c := range current {
			if c.ExposedTo != nil && c.GetExposedTo() == p.GetExposedTo() {
				return true
			}
		}
	}

	return false
}
