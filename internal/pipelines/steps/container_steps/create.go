package container_steps

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"

	"go.vervstack.ru/Velez/internal/clients/node_clients"
)

type createContainerStep struct {
	docker    node_clients.Docker
	dockerAPI client.APIClient

	req         *container.CreateRequest
	name        *string
	containerId *string
}

func Create(
	nc node_clients.NodeClients,
	req *container.CreateRequest,
	name *string,

	containerId *string,
) *createContainerStep {
	return &createContainerStep{
		docker:      nc.Docker(),
		dockerAPI:   nc.Docker().Client(),
		req:         req,
		name:        name,
		containerId: containerId,
	}
}

func (s *createContainerStep) Do(ctx context.Context) error {
	pCfg := &v1.Platform{}

	createdContainer, err := s.dockerAPI.ContainerCreate(ctx,
		s.req.Config,
		s.req.HostConfig,
		s.req.NetworkingConfig,
		pCfg,
		toolbox.FromPtr(s.name),
	)
	if err != nil {
		return rerrors.Wrap(err, "error creating container")
	}

	containerInfo, err := s.dockerAPI.ContainerInspect(ctx, createdContainer.ID)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting container by id")
	}

	*s.containerId = containerInfo.ID

	return nil
}

func (s *createContainerStep) Rollback(ctx context.Context) error {
	if s.containerId == nil {
		return nil
	}

	err := s.docker.Remove(ctx, *s.containerId)
	if err != nil {
		if !errdefs.IsNotFound(err) {
			return rerrors.Wrapf(err, "error removing container '%s'", *s.containerId)
		}
	}

	return nil
}
