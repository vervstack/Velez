package docker

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	errors "go.redsock.ru/rerrors"

	"github.com/godverv/Velez/internal/clients/docker/dockerutils"
	"github.com/godverv/Velez/pkg/velez_api"
)

type Docker struct {
	client.CommonAPIClient
}

func NewClient() (*Docker, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, errors.Wrap(err, "error getting docker client")
	}

	return &Docker{
		CommonAPIClient: cli,
	}, nil
}

func (d *Docker) PullImage(ctx context.Context, imageName string) (types.ImageInspect, error) {
	_, err := dockerutils.PullImage(ctx, d.CommonAPIClient, imageName, false)
	if err != nil {
		return types.ImageInspect{}, errors.Wrap(err, "error pulling image")
	}

	image, err := d.InspectImage(ctx, imageName)
	if err != nil {
		return types.ImageInspect{}, errors.Wrap(err, "error inspecting image")
	}

	return image, nil
}

func (d *Docker) Remove(ctx context.Context, contUUID string) error {
	err := d.ContainerRemove(ctx, contUUID,
		container.RemoveOptions{
			Force: true,
		})

	if err != nil {
		if !strings.Contains(err.Error(), NoSuchContainerError) {
			return nil
		}
		return errors.Wrap(err, "error removing container")
	}

	return nil
}

func (d *Docker) ListContainers(ctx context.Context, req *velez_api.ListSmerds_Request) ([]types.Container, error) {
	list, err := dockerutils.ListContainers(ctx, d.CommonAPIClient, req)
	if err != nil {
		return nil, errors.Wrap(err, "error listing containers")
	}

	return list, nil
}

func (d *Docker) InspectContainer(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	cont, err := d.CommonAPIClient.ContainerInspect(ctx, containerID)
	if err != nil {
		return types.ContainerJSON{}, errors.Wrap(err, "error inspecting container")
	}

	return cont, nil
}

func (d *Docker) InspectImage(ctx context.Context, image string) (types.ImageInspect, error) {
	img, _, err := d.CommonAPIClient.ImageInspectWithRaw(ctx, image)
	if err != nil {
		return img, errors.Wrap(err, "error inspecting image")
	}

	return img, nil
}
