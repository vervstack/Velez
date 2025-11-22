package smerd_steps

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients"
)

type smerdStart struct {
	dockerAPI client.APIClient

	containerId *string
}

func Start(
	nodeClients clients.NodeClients,
	containerId *string,
) *smerdStart {
	return &smerdStart{
		dockerAPI:   nodeClients.Docker().Client(),
		containerId: containerId,
	}
}

func (s *smerdStart) Do(ctx context.Context) error {
	if s.containerId == nil {
		return rerrors.New("no container id provided")
	}

	err := s.dockerAPI.ContainerStart(ctx, *s.containerId, container.StartOptions{})
	if err != nil {
		return rerrors.Wrap(err, "error starting container")
	}

	return nil
}

func (s *smerdStart) Rollback(ctx context.Context) error {
	if s.containerId == nil {
		return nil
	}

	err := s.dockerAPI.ContainerStop(ctx, *s.containerId, container.StopOptions{})
	if err != nil {
		return rerrors.Wrapf(err, "error removing container '%s'", *s.containerId)
	}

	return nil
}
