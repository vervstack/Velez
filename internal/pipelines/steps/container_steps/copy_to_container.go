package container_steps

import (
	"context"

	"github.com/docker/docker/client"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/domain"
)

type copyToContainerStep struct {
	dockerAPI client.APIClient

	contId *string
	mounts *[]domain.FileMountPoint
}

func CopyToContainer(
	nodeClients node_clients.NodeClients,
	contId *string,
	mounts *[]domain.FileMountPoint,
) *copyToContainerStep {
	return &copyToContainerStep{
		dockerAPI: nodeClients.Docker().Client(),
		contId:    contId,
		mounts:    mounts,
	}
}

func (s *copyToContainerStep) Do(ctx context.Context) error {
	for _, mount := range *s.mounts {
		m := mount
		if m.Content == nil {
			continue
		}

		err := s.validate(&m)
		if err != nil {
			return rerrors.Wrap(err, "error during validation")
		}

		err = dockerutils.WriteToContainer(ctx, s.dockerAPI,
			*s.contId, *m.FilePath,
			m.Content,
		)
		if err != nil {
			return rerrors.Wrap(err, "error copying to container")
		}
	}

	return nil
}

func (s *copyToContainerStep) validate(m *domain.FileMountPoint) error {
	if s.contId == nil {
		return rerrors.New("no container id provided")
	}

	if m.FilePath == nil {
		return rerrors.New("no file path provided")
	}

	return nil
}
