package container_steps

import (
	"context"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"

	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker"
	"go.vervstack.ru/Velez/internal/pipelines/steps"
)

type createContainerStep struct {
	dockerClient node_clients.Docker
	dockerAPI    client.APIClient

	req  *container.CreateRequest
	name *string

	containerIdResp *string

	// Flag showing that the container was created during this step execution
	// in case if container already existed - it won't be removed
	isCreated bool
}

func Create(
	nc node_clients.NodeClients,
	req *container.CreateRequest,
	name *string,

	containerIdResp *string,
) steps.Step {
	return &createContainerStep{
		dockerClient:    nc.Docker(),
		dockerAPI:       nc.Docker().Client(),
		req:             req,
		name:            name,
		containerIdResp: containerIdResp,
	}
}

func (s *createContainerStep) Do(ctx context.Context) error {
	pCfg := &v1.Platform{}

	createResp, createErr := s.dockerClient.ContainerCreate(ctx,
		s.req.Config,
		s.req.HostConfig,
		s.req.NetworkingConfig,
		pCfg,
		toolbox.FromPtr(s.name),
	)
	if createErr != nil {
		if !rerrors.Is(createErr, docker.ErrNameIsTaken) {
			return rerrors.Wrap(createErr, "error creating container")
		}
	}

	*s.containerIdResp = createResp.ID

	return createErr
}

func (s *createContainerStep) Rollback(ctx context.Context) error {
	if s.containerIdResp == nil || !s.isCreated {
		return nil
	}

	containerInfo, err := s.dockerAPI.ContainerInspect(ctx, *s.containerIdResp)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting container")
	}

	err = s.dockerClient.Remove(ctx, *s.containerIdResp)
	if err != nil {
		if !cerrdefs.IsNotFound(err) {
			return rerrors.Wrapf(err, "error removing container '%s'", *s.containerIdResp)
		}
	}

	for _, m := range containerInfo.Mounts {
		if m.Type != mount.TypeVolume {
			continue
		}
		err = s.dockerAPI.VolumeRemove(ctx, m.Name, false)
		if err != nil {
			return rerrors.Wrap(err, "error deleting volume")
		}
	}

	return nil
}
