package steps

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"go.redsock.ru/rerrors"

	"github.com/godverv/Velez/internal/clients"
)

type startContainerStep struct {
	docker clients.Docker

	containerId *string
}

func StartContainer(
	nodeClients clients.NodeClients,
	containerId *string,
) *startContainerStep {
	return &startContainerStep{
		docker:      nodeClients.Docker(),
		containerId: containerId,
	}
}

func (s *startContainerStep) Do(ctx context.Context) error {
	if s.containerId == nil {
		return rerrors.New("no container id provided")
	}

	err := s.docker.ContainerStart(ctx, *s.containerId, container.StartOptions{})
	if err != nil {
		return rerrors.Wrap(err, "error starting container")
	}

	return nil
}

func (s *startContainerStep) Rollback(ctx context.Context) error {
	if s.containerId == nil {
		return nil
	}

	err := s.docker.ContainerStop(ctx, *s.containerId, container.StopOptions{})
	if err != nil {
		return rerrors.Wrapf(err, "error removing container '%s'", *s.containerId)
	}

	return nil
}
