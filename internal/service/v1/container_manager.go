package v1

import (
	"context"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/godverv/Velez/pkg/velez_api"
)

type containerManager struct {
	docker client.CommonAPIClient
}

func NewContainerManager(docker client.CommonAPIClient) (*containerManager, error) {
	return &containerManager{
		docker: docker,
	}, nil
}

func (c *containerManager) LaunchSmerd(ctx context.Context, req *velez_api.CreateSmerd_Request) (*velez_api.Smerd, error) {
	image, err := c.getImage(ctx, req.ImageName)
	if err != nil {
		return nil, errors.Wrap(err, "error getting image")
	}

	serviceContainer, err := c.docker.ContainerCreate(ctx,
		&container.Config{
			Image:    image.Name,
			Hostname: req.Name,
			Volumes:  fromVolumes(req.Settings),
		},
		&container.HostConfig{
			PortBindings: fromPorts(req.Settings),
		},
		&network.NetworkingConfig{},
		&v1.Platform{},
		req.Name,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating container")
	}

	err = c.docker.ContainerStart(ctx, serviceContainer.ID, types.ContainerStartOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "error starting container")
	}

	out := &velez_api.Smerd{
		Uuid:      serviceContainer.ID,
		ImageName: req.ImageName,
	}

	cl, err := c.docker.ContainerList(ctx, types.ContainerListOptions{
		Filters: filters.NewArgs(filters.Arg("id", serviceContainer.ID)),
	})

	if err == nil && len(cl) == 1 && len(cl[0].Names) > 0 {
		out.Name = cl[0].Names[0][1:]
	}

	return out, nil
}
