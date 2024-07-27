package parser

import (
	"github.com/docker/docker/api/types/mount"

	"github.com/godverv/Velez/pkg/velez_api"
)

func FromBind(settings *velez_api.Container_Settings) []mount.Mount {
	if settings == nil {
		return nil
	}

	if len(settings.Volumes) == 0 {
		return nil
	}

	out := make([]mount.Mount, 0, len(settings.Volumes))

	for _, item := range settings.Volumes {
		out = append(out, mount.Mount{
			Type:   mount.TypeVolume,
			Source: item.VolumeName,
			Target: item.ContainerPath,
		})
	}

	return out
}

// TODO test out
func ToBind(volumes []mount.Mount) []*velez_api.Volume {
	if len(volumes) == 0 {
		return nil
	}

	out := make([]*velez_api.Volume, len(volumes))

	for i, item := range volumes {
		out[i] = &velez_api.Volume{
			VolumeName:    item.Source,
			ContainerPath: item.Target,
		}
	}

	return out
}
