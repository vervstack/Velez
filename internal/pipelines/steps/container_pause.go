package steps

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients"
	"go.vervstack.ru/Velez/internal/domain"
)

type pauseContainerStep struct {
	docker clients.Docker

	req         domain.LaunchSmerd
	containerId *string
}

func PauseContainer(
	nodeClients clients.NodeClients,
	containerId *string,
) *pauseContainerStep {
	return &pauseContainerStep{
		docker:      nodeClients.Docker(),
		containerId: containerId,
	}
}

func (s *pauseContainerStep) Do(ctx context.Context) error {
	if s.containerId == nil {
		return rerrors.New("container id is required")
	}

	err := s.docker.ContainerPause(ctx, *s.containerId)
	if err != nil {
		return rerrors.Wrap(err, "error pausing container")
	}

	return nil
}

func (s *pauseContainerStep) Rollback(ctx context.Context) error {
	if s.containerId == nil {
		return nil
	}

	err := s.docker.ContainerUnpause(ctx, *s.containerId)
	if err != nil {
		return rerrors.Wrapf(err, "error unpausing container '%s'", s.containerId)
	}

	return nil
}
