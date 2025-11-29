package parser

import (
	"github.com/docker/docker/api/types/mount"

	"go.vervstack.ru/Velez/pkg/velez_api"
)

func FromVolume(settings *velez_api.Container_Settings) []mount.Mount {
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

func ToVolume(volumes []mount.Mount) []*velez_api.Volume {
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
