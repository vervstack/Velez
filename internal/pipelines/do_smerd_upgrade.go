package pipelines

import (
	"context"

	"github.com/docker/docker/api/types/image"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/pipelines/steps"
	"go.vervstack.ru/Velez/internal/pipelines/steps/config_steps"
	"go.vervstack.ru/Velez/internal/pipelines/steps/smerd_steps"
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
			smerd_steps.PauseContainer(p.nodeClients, &oldContId),
			// Config stage
			steps.SingleFunc(func(_ context.Context) error {
				newLaunch.Name = req.Name + configSuffix
				return nil
			}),
			smerd_steps.Create(p.nodeClients, &newLaunch, &newContId),
			config_steps.GetConfigFromContainerStep(p.nodeClients, p.services, &newLaunch, &newContId, &img, &cfgMount),
			smerd_steps.DropContainerStep(p.nodeClients, &newContId),
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
			smerd_steps.Create(p.nodeClients, &newLaunch, &newContId),
			smerd_steps.Start(p.nodeClients, &newContId),
			steps.Healthcheck(p.nodeClients, &newLaunch, &newContId),
			// Clean up
			smerd_steps.RenameContainer(p.nodeClients, &oldContId, req.Name+"_old"),
			smerd_steps.DropContainerStep(p.nodeClients, &oldContId),
			smerd_steps.RenameContainer(p.nodeClients, &newContId, req.Name),
		},

		getResult: func() (*any, error) {
			return nil, nil
		},
	}
}
