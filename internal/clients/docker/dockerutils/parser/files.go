package parser

import (
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/mount"

	"github.com/godverv/Velez/pkg/velez_api"
)

func FromBind(settings *velez_api.Container_Settings) []mount.Mount {
	if settings == nil {
		return nil
	}

	out := make([]mount.Mount, 0, len(settings.Mounts)+len(settings.Volumes))
	for _, item := range settings.Mounts {
		out = append(out, mount.Mount{
			Type:   mount.TypeBind,
			Source: item.Host,
			Target: item.Container,
		})
	}

	for _, item := range settings.Volumes {
		out = append(out, mount.Mount{
			Type:   mount.TypeVolume,
			Source: item.Volume,
			Target: item.ContainerPath,
		})
	}

	return out
}

func ToBind(volumes []types.MountPoint) []*velez_api.MountBindings {
	out := make([]*velez_api.MountBindings, len(volumes))

	for i, item := range volumes {
		splited := strings.Split(item.Destination, ":")
		if len(splited) != 2 {
			continue
		}

		out[i] = &velez_api.MountBindings{
			Host:      splited[0],
			Container: splited[1],
		}
	}

	return out
}
