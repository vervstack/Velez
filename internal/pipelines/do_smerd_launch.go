package pipelines

import (
	"github.com/docker/docker/api/types/image"
	"go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/pipelines/steps"
	"go.vervstack.ru/Velez/internal/pipelines/steps/config_steps"
	"go.vervstack.ru/Velez/internal/pipelines/steps/container_steps"
	"go.vervstack.ru/Velez/internal/pipelines/steps/smerd_steps"
)

func (p *pipeliner) LaunchSmerd(req domain.LaunchSmerd) Runner[domain.LaunchSmerdResult] {
	if p.suffix != "" {
		req.Name = req.GetName() + "_" + p.suffix
	}

	imageResp := &image.InspectResponse{}

	containerId := ""

	mounts := make([]domain.FileMountPoint, 0)

	return &runner[domain.LaunchSmerdResult]{
		Steps: []steps.Step{
			// Prepare stage
			steps.PrepareCreateRequest(&req),
			steps.PrepareImage(p.nodeClients, req.ImageName, imageResp),
			config_steps.FetchConfig(p.services, &req, imageResp, &mounts),
			steps.PrepareVervConfig(p.nodeClients, p.services, &req, imageResp),
			// Deploy stage
			smerd_steps.Create(p.nodeClients, &req, &containerId),
			container_steps.CopyToContainer(p.nodeClients, &containerId, &mounts),
			smerd_steps.Start(p.nodeClients, &containerId),
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
