package parser

import (
	"github.com/docker/docker/api/types/mount"

	"github.com/godverv/Velez/pkg/velez_api"
)

func FromBind(settings *velez_api.Container_Settings) []mount.Mount {
	if settings == nil {
		return nil
	}

	if len(settings.Mounts) == 0 {
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

func ToBind(volumes []mount.Mount) []*velez_api.MountBindings {
	if len(volumes) == 0 {
		return nil
	}

	out := make([]*velez_api.MountBindings, len(volumes))

	for i, item := range volumes {
		out[i] = &velez_api.MountBindings{
			Host:      item.Source,
			Container: item.Target,
		}
	}

	return out
}
