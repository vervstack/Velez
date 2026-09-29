package parser

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func FromVolume(settings *velez_api.Container_Settings) []mount.Mount {
	if settings == nil {
		return nil
	}

	if len(settings.GetVolumes()) == 0 {
		return nil
	}

	out := make([]mount.Mount, 0, len(settings.GetVolumes()))

	for _, item := range settings.GetVolumes() {
		out = append(out, mount.Mount{
			Type:   mount.TypeVolume,
			Source: item.GetVolumeName(),
			Target: item.GetContainerPath(),
		})
	}

	return out
}

// ToVolume reads the daemon's resolved mounts rather than HostConfig.Mounts,
// which stays empty for containers created with binds (`docker run -v`).
// Only volume-type mounts are representable; bind mounts are skipped.
func ToVolume(mounts []container.MountPoint) []*velez_api.Volume {
	out := make([]*velez_api.Volume, 0, len(mounts))

	for _, item := range mounts {
		if item.Type != mount.TypeVolume {
			continue
		}

		out = append(out, &velez_api.Volume{
			VolumeName:    item.Name,
			ContainerPath: item.Destination,
		})
	}

	if len(out) == 0 {
		return nil
	}

	return out
}
