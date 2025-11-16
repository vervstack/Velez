package patterns

import (
	"strings"

	"github.com/docker/docker/client"

	"go.vervstack.ru/Velez/internal/domain/labels"
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
					// TODO move portainer to private service due to socket bind (not allowed)
					HostPath:      strings.ReplaceAll(client.DefaultDockerHost, "unix://", ""),
					ContainerPath: "/var/run/docker.sock",
				},
			},
		},
		Labels: map[string]string{
			labels.CreatedWithVelezLabel: "true",
		},
	}
}
