package deploy_manager

import (
	"github.com/docker/docker/api/types/container"

	"github.com/godverv/Velez/internal/clients/docker/dockerutils/parser"
	"github.com/godverv/Velez/pkg/velez_api"
)

func getHostConfig(req *velez_api.CreateSmerd_Request) (hostConfig *container.HostConfig) {
	// TODO https://redsock.youtrack.cloud/issue/VERV-56
	hostConfig = &container.HostConfig{
		PortBindings: parser.FromPorts(req.Settings),
		Mounts:       parser.FromBind(req.Settings),
	}

	if req.Settings != nil && len(req.Settings.Volumes) != 0 {
		hostConfig.VolumeDriver = req.Settings.Volumes[0].VolumeName
	}

	return hostConfig
}
