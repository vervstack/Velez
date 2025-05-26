package steps

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients"
)

type renameContainerStep struct {
	docker clients.Docker

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
		docker:      nodeClients.Docker(),
		containerId: containerId,
		newName:     newName,
	}
}

func (s *renameContainerStep) Do(ctx context.Context) error {
	if s.containerId == nil {
		return rerrors.New("container id is required")
	}

	smerd, err := s.docker.InspectContainer(ctx, *s.containerId)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting container")
	}

	s.oldName = smerd.Name

	err = s.docker.ContainerRename(ctx, *s.containerId, s.newName)
	if err != nil {
		return rerrors.Wrap(err, "error renaming container")
	}

	return nil
}

func (s *renameContainerStep) Rollback(ctx context.Context) error {
	err := s.docker.ContainerRename(ctx, *s.containerId, s.oldName)
	if err != nil {
		return rerrors.Wrap(err, "error renaming container on rollback")
	}

	return nil
}
