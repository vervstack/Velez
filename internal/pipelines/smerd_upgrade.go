package pipelines

import (
	"context"

	"github.com/docker/docker/api/types/image"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/pipelines/steps"
)

func (p *pipeliner) UpgradeSmerd(req domain.UpgradeSmerd) Runner[any] {
	newLaunch := domain.LaunchSmerd{}
	img := image.InspectResponse{}
	var newContId string
	var oldContId string
	appConfig := domain.AppConfig{}

	return &runner[any]{
		Steps: []steps.Step{
			// Preparation stage
			steps.FromContainerToRequest(req.Name, p.services.SmerdManager(), &newLaunch, &oldContId),
			steps.PrepareImageStep(p.nodeClients, req.Image, &img),
			steps.SingleFunc(func(_ context.Context) error {
				newLaunch.ImageName = req.Image
				newLaunch.Name = req.Name + "_new"
				return nil
			}),
			steps.PrepareVervConfig(p.nodeClients, p.services, &newLaunch, &img),
			steps.PauseContainer(p.nodeClients, &oldContId),
			// Deploy stage
			steps.CreateContainer(p.nodeClients, &newLaunch, &newContId),
			steps.AssembleConfigStep(p.nodeClients, p.services, &newContId, &newLaunch, &img, &appConfig),
			steps.StartSmerd(p.nodeClients, &newContId),
			steps.HealthcheckStep(p.nodeClients, &newLaunch, &newContId),
			// Clean up
			steps.RenameContainer(p.nodeClients, &oldContId, req.Name+"_old"),
			steps.DropContainerStep(p.nodeClients, &oldContId),
			steps.RenameContainer(p.nodeClients, &newContId, req.Name),
		},

		getResult: func() (*any, error) {
			return nil, nil
		},
	}
}
