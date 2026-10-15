package transfer_out

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
)

func bindMountSources(inspected container.InspectResponse) []string {
	sources := make([]string, 0)

	for _, mnt := range inspected.Mounts {
		if mnt.Type == mount.TypeBind {
			sources = append(sources, mnt.Source)
		}
	}

	return sources
}

func hasBindMounts(inspected container.InspectResponse) bool {
	return len(bindMountSources(inspected)) > 0
}
