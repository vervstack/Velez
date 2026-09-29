package container_manager

import (
	"sort"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

// toMounts converts Docker's runtime mount points (shared shape between
// container.Summary.Mounts and container.InspectResponse.Mounts) into the
// DockerContainer.Mounts wire shape.
func toMounts(mounts []container.MountPoint) []*velez_api.Mount {
	out := make([]*velez_api.Mount, 0, len(mounts))

	for _, m := range mounts {
		out = append(out, &velez_api.Mount{
			Source:      m.Source,
			Destination: m.Destination,
			Type:        string(m.Type),
			ReadWrite:   m.RW,
		})
	}

	return out
}

// toNetworkBinds converts Docker's per-network endpoint settings (shared
// shape between container.Summary.NetworkSettings.Networks and
// container.InspectResponse.NetworkSettings.Networks) into
// DockerContainer.Networks, sorted by network name for a stable response -
// same convention as InspectSmerd's equivalent Smerd.Networks loop in
// inspect.go.
func toNetworkBinds(containerID string, networks map[string]*network.EndpointSettings) []*velez_api.NetworkBind {
	out := make([]*velez_api.NetworkBind, 0, len(networks))

	for netName, net := range networks {
		nb := &velez_api.NetworkBind{
			NetworkName: netName,
		}

		if net != nil {
			ip := net.IPAddress

			nb.IpAddress = &ip

			if len(net.DNSNames) != 0 {
				nb.Aliases = make([]string, 0, len(net.DNSNames)-1)
			}

			for _, dName := range net.DNSNames {
				if !strings.HasPrefix(containerID, dName) {
					nb.Aliases = append(nb.Aliases, dName)
				}
			}
		}

		out = append(out, nb)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].GetNetworkName() < out[j].GetNetworkName()
	})

	return out
}

// matchesContainerFilters reports whether dc satisfies every filter in
// filters (AND-combined, empty/nil filters always matches).
func matchesContainerFilters(dc *velez_api.DockerContainer, filters []*velez_api.ContainerFilter) bool {
	for _, f := range filters {
		if !matchesContainerFilter(dc, f) {
			return false
		}
	}

	return true
}

// matchesContainerFilter dispatches on f.Field. An unrecognized field
// (including the zero value) matches everything rather than erroring - see
// ContainerFilterField's doc comment in velez_common.proto for why.
func matchesContainerFilter(dc *velez_api.DockerContainer, f *velez_api.ContainerFilter) bool {
	switch f.GetField() {
	case velez_api.ContainerFilterField_service:
		return matchesServiceFilter(dc, f.GetValue())
	default:
		return true
	}
}

// matchesServiceFilter reports whether value is a case-insensitive
// substring of dc's resolved display name - velez.display_name when the
// container carries it, else the raw VERV_SERVICE label, mirroring the
// frontend's own deriveServiceDisplayName fallback
// (pkg/web/Velez-UI/src/processes/mappings/smerds.ts).
func matchesServiceFilter(dc *velez_api.DockerContainer, value string) bool {
	name := dc.GetLabels()[labels.DisplayNameLabel]
	if name == "" {
		name = dc.GetLabels()[labels.VervServiceLabel]
	}

	if name == "" {
		return false
	}

	return strings.Contains(strings.ToLower(name), strings.ToLower(value))
}
