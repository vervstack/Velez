package patterns

import (
	"github.com/docker/docker/client"

	"go.vervstack.ru/Velez/pkg/velez_api"
)

const (
	PortainerServiceName = "portainer"
	PortainerImage       = "portainer/portainer-ce:2.33.2-alpine"
	PortainerPort        = 9000
)

func Portainer() *velez_api.CreateSmerd_Request {

	return &velez_api.CreateSmerd_Request{
		Name:      PortainerServiceName,
		ImageName: PortainerImage,
		Hardware: &velez_api.Container_Hardware{
			CpuAmount:    nil,
			RamMb:        nil,
			MemorySwapMb: nil,
		},
		Settings: &velez_api.Container_Settings{
			Ports: []*velez_api.Port{
				{
					ServicePortNumber: PortainerPort,
				},
			},
			Volumes: []*velez_api.Volume{
				{
					VolumeName:    "portainer_data",
					ContainerPath: "/data",
				},
			},
			Binds: []*velez_api.Bind{
				{
					HostPath:      client.DefaultDockerHost,
					ContainerPath: "/var/run/docker.sock",
				},
			},
		},
	}
}
