package pipelines

import (
	"context"

	"github.com/docker/docker/api/types/image"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/pipelines/steps"
	"go.vervstack.ru/Velez/internal/pipelines/steps/config_steps"
	"go.vervstack.ru/Velez/internal/pipelines/steps/container_steps"
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
			container_steps.PauseContainer(p.nodeClients, &oldContId),
			// Config stage
			steps.SingleFunc(func(_ context.Context) error {
				newLaunch.Name = req.Name + configSuffix
				return nil
			}),
			container_steps.CreateContainer(p.nodeClients, &newLaunch, &newContId),
			config_steps.GetConfigFromContainerStep(p.nodeClients, p.services, &newLaunch, &newContId, &img, &cfgMount),
			container_steps.DropContainerStep(p.nodeClients, &newContId),
			// Config stage
			steps.SingleFunc(func(_ context.Context) error {
				newLaunch.Name = req.Name
				return nil
			}),
			// Deploy stage
			config_steps.FetchConfig(p.services, &newLaunch, &img, &cfgMount),
			steps.PrepareVervConfig(p.nodeClients, p.services, &newLaunch, &img),
			// Config stage
			steps.SingleFunc(func(_ context.Context) error {
				newLaunch.Name = req.Name + "_new"
				return nil
			}),
			container_steps.CreateContainer(p.nodeClients, &newLaunch, &newContId),
			container_steps.StartSmerd(p.nodeClients, &newContId),
			steps.Healthcheck(p.nodeClients, &newLaunch, &newContId),
			// Clean up
			container_steps.RenameContainer(p.nodeClients, &oldContId, req.Name+"_old"),
			container_steps.DropContainerStep(p.nodeClients, &oldContId),
			container_steps.RenameContainer(p.nodeClients, &newContId, req.Name),
		},

		getResult: func() (*any, error) {
			return nil, nil
		},
	}
}
