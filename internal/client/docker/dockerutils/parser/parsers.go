package parser

import (
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/go-connections/nat"

	"github.com/godverv/Velez/pkg/velez_api"
)

func FromVolumes(settings *velez_api.Container_Settings) map[string]struct{} {
	if settings == nil {
		return nil
	}

	out := map[string]struct{}{}
	for _, item := range settings.Volumes {
		out[item.Host+":"+item.Container] = struct{}{}
	}

	return out
}

func FromPorts(settings *velez_api.Container_Settings) map[nat.Port][]nat.PortBinding {
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

func FromCommand(command *string) strslice.StrSlice {
	if command == nil {
		return nil
	}

	return strings.Split(*command, " ")
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

func ToPorts(ports []types.Port) []*velez_api.PortBindings {
	out := make([]*velez_api.PortBindings, len(ports))

	for i, item := range ports {
		out[i] = &velez_api.PortBindings{
			Host:      uint32(item.PublicPort),
			Container: uint32(item.PrivatePort),
			Protoc:    velez_api.PortBindings_Protocol(velez_api.PortBindings_Protocol_value[item.Type]),
		}
	}
	return out
}
