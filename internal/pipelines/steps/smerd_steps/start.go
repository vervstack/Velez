package smerd_steps

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
)

type smerdStart struct {
	dockerAPI client.APIClient

	containerID *string
}

func Start(
	nodeClients node_clients.NodeClients,
	containerID *string,
) *smerdStart {
	return &smerdStart{
		dockerAPI:   nodeClients.Docker().Client(),
		containerID: containerID,
	}
}

func (s *smerdStart) Do(ctx context.Context) error {
	if s.containerID == nil {
		return rerrors.New("no container id provided")
	}

	err := s.dockerAPI.ContainerStart(ctx, *s.containerID, container.StartOptions{})
	if err != nil {
		return rerrors.Wrap(err, "error starting container")
	}

	return nil
}

func (s *smerdStart) Rollback(ctx context.Context) error {
	if s.containerID == nil {
		return nil
	}

	err := s.dockerAPI.ContainerStop(ctx, *s.containerID, container.StopOptions{})
	if err != nil {
		return rerrors.Wrapf(err, "error removing container '%s'", *s.containerID)
	}

	return nil
}
