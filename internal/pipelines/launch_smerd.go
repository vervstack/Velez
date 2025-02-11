package pipelines

import (
	"github.com/docker/docker/api/types"
	"go.redsock.ru/rerrors"

	"github.com/godverv/Velez/internal/domain"
	"github.com/godverv/Velez/internal/pipelines/steps"
)

func (p *pipeliner) LaunchSmerd(req domain.LaunchSmerd) Runner[domain.LaunchSmerdResult] {
	image := &types.ImageInspect{}

	containerId := ""

	return &runner[domain.LaunchSmerdResult]{
		Steps: []steps.Step{
			// Prepare steps
			steps.PrepareCreateRequest(req.CreateSmerd_Request),
			steps.PrepareImageStep(p.dockerAPI, req.ImageName, image),
			steps.PrepareVervConfig(p.configService, p.portManager, req.CreateSmerd_Request, image),
			// Deploy steps
			steps.CreateContainer(p.dockerAPI, req.CreateSmerd_Request, &containerId),
			steps.StartContainer(p.dockerAPI, &containerId),
			// Post deploy steps
			steps.HealthcheckStep(p.dockerAPI, req.CreateSmerd_Request, &containerId),
			steps.SubscribeForConfigChanges(p.configService, req.CreateSmerd_Request),
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
