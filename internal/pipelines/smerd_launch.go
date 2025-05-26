package pipelines

import (
	"github.com/docker/docker/api/types/image"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/pipelines/steps"
)

func (p *pipeliner) LaunchSmerd(req domain.LaunchSmerd) Runner[domain.LaunchSmerdResult] {
	imageResp := &image.InspectResponse{}

	containerId := ""

	cfg := &domain.AppConfig{}

	return &runner[domain.LaunchSmerdResult]{
		Steps: []steps.Step{
			// Prepare stage
			steps.PrepareCreateRequest(&req),
			steps.PrepareImageStep(p.nodeClients, req.ImageName, imageResp),
			steps.PrepareVervConfig(p.nodeClients, p.services, &req, imageResp),
			// Deploy stage
			steps.CreateContainer(p.nodeClients, &req, &containerId),
			steps.AssembleConfigStep(p.nodeClients, p.services, &containerId, &req, imageResp, cfg),
			steps.StartSmerd(p.nodeClients, &containerId),
			// Post deploy stage
			steps.HealthcheckStep(p.nodeClients, &req, &containerId),
			steps.SubscribeForConfigChanges(p.services, req.Name),
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
