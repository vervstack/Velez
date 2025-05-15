package pipelines

import (
	"github.com/docker/docker/api/types"
	"go.redsock.ru/rerrors"
	"go.vervstack.ru/matreshka/pkg/matreshka"

	"github.com/godverv/Velez/internal/domain"
	"github.com/godverv/Velez/internal/pipelines/steps"
)

func (p *pipeliner) LaunchSmerd(req domain.LaunchSmerd) Runner[domain.LaunchSmerdResult] {
	image := &types.ImageInspect{}

	containerId := ""

	cfg := &matreshka.AppConfig{}

	return &runner[domain.LaunchSmerdResult]{
		Steps: []steps.Step{
			// Prepare steps
			steps.PrepareCreateRequest(req.CreateSmerd_Request),
			steps.PrepareImageStep(p.nodeClients, req.ImageName, image),
			steps.PrepareVervConfig(p.nodeClients, p.services, req.CreateSmerd_Request, image),
			// Deploy steps
			steps.CreateContainer(p.nodeClients, req, &containerId),
			steps.AssembleConfigStep(p.nodeClients, p.services, &containerId, req, image, cfg),
			steps.StartContainer(p.nodeClients, &containerId),
			// Post deploy steps
			steps.HealthcheckStep(p.nodeClients, req.CreateSmerd_Request, &containerId),
			steps.SubscribeForConfigChanges(p.services, req.CreateSmerd_Request),
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
