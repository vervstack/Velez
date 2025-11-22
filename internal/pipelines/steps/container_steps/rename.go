package container_steps

import (
	"context"

	"github.com/docker/docker/client"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients"
)

type renameContainerStep struct {
	dockerAPI client.APIClient

	containerId *string
	newName     string

	oldName string
}

func RenameContainer(
	nodeClients clients.NodeClients,
	containerId *string,
	newName string,
) *renameContainerStep {
	return &renameContainerStep{
		dockerAPI:   nodeClients.Docker().Client(),
		containerId: containerId,
		newName:     newName,
	}
}

func (s *renameContainerStep) Do(ctx context.Context) error {
	if s.containerId == nil {
		return rerrors.New("container id is required")
	}

	smerd, err := s.dockerAPI.ContainerInspect(ctx, *s.containerId)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting container")
	}

	s.oldName = smerd.Name

	err = s.dockerAPI.ContainerRename(ctx, *s.containerId, s.newName)
	if err != nil {
		return rerrors.Wrap(err, "error renaming container")
	}

	return nil
}

func (s *renameContainerStep) Rollback(ctx context.Context) error {
	err := s.dockerAPI.ContainerRename(ctx, *s.containerId, s.oldName)
	if err != nil {
		return rerrors.Wrap(err, "error renaming container on rollback")
	}

	return nil
}
