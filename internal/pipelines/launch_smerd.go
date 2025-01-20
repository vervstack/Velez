package pipelines

import (
	"go.redsock.ru/rerrors"

	"github.com/godverv/Velez/internal/domain"
	"github.com/godverv/Velez/internal/pipelines/deploy_steps"
)

func (p *pipeliner) LaunchSmerd(req domain.LaunchSmerd) Runner[domain.LaunchSmerdResult] {
	deploymentState := &domain.LaunchSmerdState{}
	return &runner[domain.LaunchSmerdResult]{
		Steps: []PipelineStep{
			// Prepare steps
			deploy_steps.PrepareRequest(req.CreateSmerd_Request),
			deploy_steps.PrepareImageStep(p.dockerAPI, req.ImageName, deploymentState),
			deploy_steps.PrepareVervConfig(p.configService, req.CreateSmerd_Request, deploymentState),
			// Deploy steps
			deploy_steps.LaunchContainer(p.dockerAPI, req.CreateSmerd_Request, deploymentState),
			// Post deploy steps
			deploy_steps.HealthcheckStep(p.dockerAPI, req.CreateSmerd_Request, deploymentState),
			deploy_steps.SubscribeForConfigChanges(p.configService, req.CreateSmerd_Request),
		},

		getResult: func() (*domain.LaunchSmerdResult, error) {
			if deploymentState.ContainerId == nil {
				return nil, rerrors.New("pipeline didn't return any created container")
			}

			return &domain.LaunchSmerdResult{
				ContainerId: *deploymentState.ContainerId,
			}, nil
		},
	}
}
