package v1

import (
	"context"
	"path"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/godverv/Velez/internal/domain"
)

type containerManager struct {
	docker client.CommonAPIClient
}

func NewContainerManager(docker client.CommonAPIClient) (*containerManager, error) {
	return &containerManager{
		docker: docker,
	}, nil
}

func (c *containerManager) CreateAndRun(ctx context.Context, req domain.ContainerCreate) (domain.Container, error) {
	image, err := c.getImage(ctx, req.ImageName)
	if err != nil {
		return domain.Container{}, errors.Wrap(err, "error getting image")
	}

	baseName := path.Base(req.ImageName)
	serviceContainer, err := c.docker.ContainerCreate(ctx,
		&container.Config{
			Image:    image.Name,
			Hostname: baseName,
		},
		&container.HostConfig{},
		&network.NetworkingConfig{},
		&v1.Platform{},
		baseName,
	)
	if err != nil {
		return domain.Container{}, errors.Wrap(err, "error creating container")
	}

	err = c.docker.ContainerStart(ctx, serviceContainer.ID, types.ContainerStartOptions{})
	if err != nil {
		return domain.Container{}, errors.Wrap(err, "error starting container")
	}

	return domain.Container{}, nil
}
