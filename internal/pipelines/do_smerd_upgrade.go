package pipelines

import (
	"context"

	"github.com/docker/docker/api/types/image"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/pipelines/steps"
)

const configSuffix = "_configuration_fetcher"

func (p *pipeliner) UpgradeSmerd(req domain.UpgradeSmerd) Runner[any] {
	newLaunch := domain.LaunchSmerd{}
	img := image.InspectResponse{}
	var newContId string
	var oldContId string

	var cfgMount domain.ConfigMount

	return &runner[any]{
		Steps: []steps.Step{
			// Preparation stage
			steps.FromContainerToRequest(req.Name, p.services.SmerdManager(), &newLaunch, &oldContId),
			steps.PrepareImage(p.nodeClients, req.Image, &img),
			steps.SingleFunc(func(_ context.Context) error {
				newLaunch.ImageName = req.Image
				return nil
			}),
			steps.PauseContainer(p.nodeClients, &oldContId),
			// Config stage
			steps.SingleFunc(func(_ context.Context) error {
				newLaunch.Name = req.Name + configSuffix
				return nil
			}),
			steps.CreateContainer(p.nodeClients, &newLaunch, &newContId),
			steps.GetConfigFromContainerStep(p.nodeClients, p.services, &newLaunch, &newContId, &img, &cfgMount),
			steps.DropContainerStep(p.nodeClients, &newContId),
			// Config stage
			steps.SingleFunc(func(_ context.Context) error {
				newLaunch.Name = req.Name
				return nil
			}),
			// Deploy stage
			steps.FetchConfig(p.services, &newLaunch, &img, &cfgMount),
			steps.PrepareVervConfig(p.nodeClients, p.services, &newLaunch, &img),
			// Config stage
			steps.SingleFunc(func(_ context.Context) error {
				newLaunch.Name = req.Name + "_new"
				return nil
			}),
			steps.CreateContainer(p.nodeClients, &newLaunch, &newContId),
			steps.StartSmerd(p.nodeClients, &newContId),
			steps.Healthcheck(p.nodeClients, &newLaunch, &newContId),
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
