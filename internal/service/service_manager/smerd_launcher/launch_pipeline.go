package smerd_launcher

import (
	"context"
	"errors"

	"go.redsock.ru/rerrors"

	"github.com/godverv/Velez/internal/clients"
	"github.com/godverv/Velez/internal/service"
	"github.com/godverv/Velez/internal/service/service_manager/smerd_launcher/shared"
	"github.com/godverv/Velez/internal/service/service_manager/smerd_launcher/steps"
	"github.com/godverv/Velez/pkg/velez_api"
)

type SmerdLauncher struct {
	docker        clients.Docker
	portManager   clients.PortManager
	configService service.ConfigurationService
}

func New(
	nodeClients clients.NodeClients,
	configService service.ConfigurationService,
) *SmerdLauncher {
	return &SmerdLauncher{
		docker:      nodeClients.Docker(),
		portManager: nodeClients.PortManager(),

		configService: configService,
	}
}

type LaunchPipeline struct {
	Steps []shared.Step
}

func (s *SmerdLauncher) LaunchSmerd(ctx context.Context, req *velez_api.CreateSmerd_Request) (string, error) {
	deploymentProcess := &shared.DeployProcess{}
	pipeline := LaunchPipeline{
		Steps: []shared.Step{
			// Prepare steps
			steps.PrepareRequest(req),
			steps.PrepareImageStep(s.docker, req, deploymentProcess),
			steps.PrepareVervConfig(s.configService, req, deploymentProcess),
			// Deploy steps
			steps.LaunchContainer(s.docker, req, deploymentProcess),
			// Post deploy steps
			steps.HealthcheckStep(s.docker, req, deploymentProcess),
			steps.SubscribeForConfigChanges(req, s.configService),
		},
	}

	err := pipeline.run(ctx)
	if err != nil {
		rollbackErr := pipeline.rollback(ctx)
		if rollbackErr != nil {
			err = errors.Join(err, rerrors.Wrap(rollbackErr, "error during rolling back"))
		}

		return "", err
	}

	return deploymentProcess.Container.ID, nil
}

func (p *LaunchPipeline) run(ctx context.Context) error {
	for _, step := range p.Steps {
		err := step.Do(ctx)
		if err != nil {
			return rerrors.Wrapf(err, "error during execution of step: %T", step)
		}
	}

	return nil
}

func (p *LaunchPipeline) rollback(ctx context.Context) error {
	var globalErr error
	for _, step := range p.Steps {
		rollbackable, ok := step.(shared.RollbackableStep)
		if ok {
			err := rollbackable.Rollback(ctx)
			if err != nil {
				globalErr = errors.Join(globalErr, rerrors.Wrap(err, "error during rollback step: %v ", rollbackable))
			}
		}
	}

	return globalErr
}
