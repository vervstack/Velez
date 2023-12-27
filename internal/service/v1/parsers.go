package v1

import (
	"strconv"

	"github.com/docker/go-connections/nat"

	"github.com/godverv/Velez/pkg/velez_api"
)

func fromVolumes(settings *velez_api.Container_Settings) map[string]struct{} {
	if settings == nil {
		return nil
	}

	out := map[string]struct{}{}
	for _, item := range settings.Volumes {
		out[item.Host+":"+item.Container] = struct{}{}
	}

	return out
}

func fromPorts(settings *velez_api.Container_Settings) map[nat.Port][]nat.PortBinding {
	if settings == nil {
		return nil
	}

	out := make(map[nat.Port][]nat.PortBinding, len(settings.Ports))

	for _, item := range settings.Ports {
		containerPort, _ := nat.NewPort(item.Protoc.String(), strconv.FormatUint(uint64(item.Container), 10))

		out[containerPort] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: strconv.FormatUint(uint64(item.Host), 10),
			},
		}
	}

	return out
}
