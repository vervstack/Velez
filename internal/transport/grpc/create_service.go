package grpc

import (
	"context"

	"github.com/docker/docker/api/types/container"

	"github.com/godverv/Velez/pkg/velez_api"
)

func (a *Api) CreateService(ctx context.Context, req *velez_api.CreateService_Request) (*velez_api.CreateService_Response, error) {
	a.dockerAPI.ContainerCreate(ctx,
		&container.Config{
			Image: req.ImageName,
		},
		&container.HostConfig{
			Binds:           nil,
			ContainerIDFile: "",
			LogConfig:       container.LogConfig{},
			NetworkMode:     "",
			PortBindings:    nil,
			RestartPolicy:   container.RestartPolicy{},
			AutoRemove:      false,
			VolumeDriver:    "",
			VolumesFrom:     nil,
			ConsoleSize:     [2]uint{},
			Annotations:     nil,
			CapAdd:          nil,
			CapDrop:         nil,
			CgroupnsMode:    "",
			DNS:             nil,
			DNSOptions:      nil,
			DNSSearch:       nil,
			ExtraHosts:      nil,
			GroupAdd:        nil,
			IpcMode:         "",
			Cgroup:          "",
			Links:           nil,
			OomScoreAdj:     0,
			PidMode:         "",
			Privileged:      false,
			PublishAllPorts: false,
			ReadonlyRootfs:  false,
			SecurityOpt:     nil,
			StorageOpt:      nil,
			Tmpfs:           nil,
			UTSMode:         "",
			UsernsMode:      "",
			ShmSize:         0,
			Sysctls:         nil,
			Runtime:         "",
			Isolation:       "",
			Resources:       container.Resources{},
			Mounts:          nil,
			MaskedPaths:     nil,
			ReadonlyPaths:   nil,
			Init:            nil,
		}, nil, nil, "",
	)
	return &velez_api.CreateService_Response{}, nil
}
