package steps

import (
	"context"

	"github.com/docker/docker/client"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients"
	"go.vervstack.ru/Velez/internal/clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/domain"
)

type mountConfigStep struct {
	dockerAPI client.APIClient

	contId *string
	mount  *domain.ConfigMount
}

func MountConfig(
	nodeClients clients.NodeClients,
	contId *string,
	mount *domain.ConfigMount,
) *mountConfigStep {
	return &mountConfigStep{
		dockerAPI: nodeClients.Docker().Client(),
		contId:    contId,
		mount:     mount,
	}
}

func (s *mountConfigStep) Do(ctx context.Context) error {
	if s.mount == nil {
		// nothing to mount
		return nil
	}

	if s.mount.Content == nil {
		// no actual file-content. May be set via env
		return nil
	}

	err := s.validate()
	if err != nil {
		return rerrors.Wrap(err, "error during validation")
	}

	err = dockerutils.WriteToContainer(ctx, s.dockerAPI,
		*s.contId, *s.mount.FilePath,
		s.mount.Content,
	)
	if err != nil {
		return rerrors.Wrap(err, "error copying to container")
	}

	return nil
}

func (s *mountConfigStep) validate() error {
	if s.contId == nil {
		return rerrors.New("no container id provided")
	}

	if s.mount.FilePath == nil {
		return rerrors.New("no file path provided")
	}

	return nil

}
