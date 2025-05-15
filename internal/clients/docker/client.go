package docker

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	errors "go.redsock.ru/rerrors"

	"github.com/godverv/Velez/internal/clients/docker/dockerutils"
	"github.com/godverv/Velez/pkg/velez_api"
)

type Docker struct {
	client.APIClient
}

func NewClient() (*Docker, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, errors.Wrap(err, "error getting docker client")
	}

	return &Docker{
		APIClient: cli,
	}, nil
}

func (d *Docker) PullImage(ctx context.Context, imageName string) (image.InspectResponse, error) {
	_, err := dockerutils.PullImage(ctx, d.APIClient, imageName, false)
	if err != nil {
		return image.InspectResponse{}, errors.Wrap(err, "error pulling image")
	}

	img, err := d.InspectImage(ctx, imageName)
	if err != nil {
		return image.InspectResponse{}, errors.Wrap(err, "error inspecting image")
	}

	return img, nil
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

func (d *Docker) ListContainers(ctx context.Context, req *velez_api.ListSmerds_Request) ([]container.Summary, error) {
	list, err := dockerutils.ListContainers(ctx, d.APIClient, req)
	if err != nil {
		return nil, errors.Wrap(err, "error listing containers")
	}

	return list, nil
}

func (d *Docker) InspectContainer(ctx context.Context, containerID string) (container.InspectResponse, error) {
	cont, err := d.APIClient.ContainerInspect(ctx, containerID)
	if err != nil {
		return container.InspectResponse{}, errors.Wrap(err, "error inspecting container")
	}

	return cont, nil
}

func (d *Docker) InspectImage(ctx context.Context, image string) (image.InspectResponse, error) {
	img, _, err := d.APIClient.ImageInspectWithRaw(ctx, image)
	if err != nil {
		return img, errors.Wrap(err, "error inspecting image")
	}

	return img, nil
}
