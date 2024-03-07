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

	out := make([]mount.Mount, 0, len(settings.Volumes))
	for _, item := range settings.Volumes {
		out = append(out, mount.Mount{
			Type:   "bind",
			Source: item.Host,
			Target: item.Container,
		})
	}

	return out
}

func ToVolumes(volumes []types.MountPoint) []*velez_api.VolumeBindings {
	out := make([]*velez_api.VolumeBindings, len(volumes))

	for i, item := range volumes {
		splited := strings.Split(item.Destination, ":")
		if len(splited) != 2 {
			continue
		}

		out[i] = &velez_api.VolumeBindings{
			Host:      splited[0],
			Container: splited[1],
		}
	}

	return out
}
