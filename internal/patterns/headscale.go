package patterns

import (
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

const (
	HeadScaleServiceName = "headscale"
	HeadScaleImage       = "docker.io/headscale/headscale:v0"
	HeadScalePort        = 8080
)

func HeadScale() *velez_api.CreateSmerd_Request {
	return &velez_api.CreateSmerd_Request{
		Name:      HeadScaleServiceName,
		ImageName: HeadScaleImage,
		Settings: &velez_api.Container_Settings{
			Ports: []*velez_api.Port{
				{
					ServicePortNumber: HeadScalePort,
				},
			},
			Volumes: []*velez_api.Volume{},
			Binds:   []*velez_api.Bind{},
		},
		Labels: map[string]string{
			labels.CreatedWithVelezLabel: "true",
		},
	}
}
