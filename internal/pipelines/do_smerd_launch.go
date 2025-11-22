package pipelines

import (
	"github.com/docker/docker/api/types/image"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/pipelines/steps"
	"go.vervstack.ru/Velez/internal/pipelines/steps/config_steps"
	"go.vervstack.ru/Velez/internal/pipelines/steps/container_steps"
)

func (p *pipeliner) LaunchSmerd(req domain.LaunchSmerd) Runner[domain.LaunchSmerdResult] {
	imageResp := &image.InspectResponse{}

	containerId := ""

	cfgMount := &domain.ConfigMount{}

	return &runner[domain.LaunchSmerdResult]{
		Steps: []steps.Step{
			// Prepare stage
			steps.PrepareCreateRequest(&req),
			steps.PrepareImage(p.nodeClients, req.ImageName, imageResp),
			config_steps.FetchConfig(p.services, &req, imageResp, cfgMount),
			steps.PrepareVervConfig(p.nodeClients, p.services, &req, imageResp),
			// Deploy stage
			container_steps.CreateContainer(p.nodeClients, &req, &containerId),
			config_steps.MountConfig(p.nodeClients, &containerId, cfgMount),
			container_steps.StartSmerd(p.nodeClients, &containerId),
			// Post deploy stage
			steps.Healthcheck(p.nodeClients, &req, &containerId),
			config_steps.SubscribeForConfigChanges(p.services, req.Name),
		},

		getResult: func() (*domain.LaunchSmerdResult, error) {
			if containerId == "" {
				return nil, rerrors.New("pipeline didn't return any created container")
			}

			return &domain.LaunchSmerdResult{
				ContainerId: containerId,
			}, nil
		},
	}
}
