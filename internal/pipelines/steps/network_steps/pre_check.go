package network_steps

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/pipelines/steps"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

type checkSidecarExist struct {
	docker      node_clients.Docker
	sideCarName string
}

func CheckSidecarExist(nc node_clients.NodeClients, sideCarName string) steps.Step {
	return &checkSidecarExist{
		docker:      nc.Docker(),
		sideCarName: sideCarName,
	}
}

func (s *checkSidecarExist) Do(ctx context.Context) error {
	// Check if there is a headscale to connect to
	r := &velez_api.ListSmerds_Request{
		Name: &s.sideCarName,
	}

	conts, err := s.docker.ListContainers(ctx, r)
	if err != nil {
		return rerrors.Wrap(err, "error listing container")
	}

	if len(conts) != 0 {
		return rerrors.Wrap(steps.ErrAlreadyExists)
	}

	return nil
}
