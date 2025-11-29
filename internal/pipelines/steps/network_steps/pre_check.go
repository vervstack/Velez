package network_steps

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/pipelines/steps"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

type preCheck struct {
	docker      node_clients.Docker
	sideCarName string
}

func PreCheck(services service.Services, sideCarName string) steps.Step {
	return &preCheck{
		docker:      services.Docker(),
		sideCarName: sideCarName,
	}
}

func (g *preCheck) Do(ctx context.Context) error {
	// Check if there is a headscale to connect to
	r := &velez_api.ListSmerds_Request{
		Name: &g.sideCarName,
	}

	conts, err := g.docker.ListContainers(ctx, r)
	if err != nil {
		return rerrors.Wrap(err, "error listing container")
	}

	_ = conts

	return nil
}
