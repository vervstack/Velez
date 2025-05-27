package parser

import (
	"strings"

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

func FromBinds(settings *velez_api.Container_Settings) []string {
	if settings == nil {
		return nil
	}

	out := make([]string, 0, len(settings.Binds))

	for _, item := range settings.Binds {
		out = append(out, item.HostPath+":"+item.ContainerPath)
	}

	return out
}

func ToBinds(binds []string) []*velez_api.Bind {
	if len(binds) == 0 {
		return nil
	}

	out := make([]*velez_api.Bind, len(binds))

	for i, item := range binds {
		separated := strings.Split(item, ":")
		if len(separated) != 2 {
			continue
		}

		out[i] = &velez_api.Bind{
			HostPath:      separated[0],
			ContainerPath: separated[1],
		}
	}

	return out
}
