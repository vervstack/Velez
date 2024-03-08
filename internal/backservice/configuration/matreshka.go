package configuration

import (
	"context"
	"time"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/godverv/Velez/internal/backservice/env"
	"github.com/godverv/Velez/internal/clients/docker/dockerutils"
	"github.com/godverv/Velez/internal/domain"
	"github.com/godverv/Velez/pkg/velez_api"
)

const (
	containerName = "matreshka"
	image         = "godverv/matreshka-be"
	duration      = time.Second * 5
)

type Matreshka struct {
	dockerAPI client.CommonAPIClient

	duration time.Duration
}

func New(dockerAPI client.CommonAPIClient) *Matreshka {
	w := &Matreshka{
		dockerAPI: dockerAPI,
		duration:  duration,
	}

	return w
}

func (b *Matreshka) Start() error {
	isAlive, err := b.IsAlive()
	if err != nil || isAlive {
		return err
	}

	ctx := context.Background()

	_, err = dockerutils.PullImage(ctx, b.dockerAPI, domain.ImageListRequest{
		Name: image,
	})
	if err != nil {
		return errors.Wrap(err, "error pulling matreshka image")
	}

	cont, err := b.dockerAPI.ContainerCreate(ctx,
		&container.Config{
			Image:    image,
			Hostname: containerName,
		},
		&container.HostConfig{},
		&network.NetworkingConfig{},
		&v1.Platform{},
		containerName,
	)
	if err != nil {
		return errors.Wrap(err, "error creating matreshka container")
	}

	err = b.dockerAPI.ContainerStart(ctx, cont.ID, types.ContainerStartOptions{})
	if err != nil {
		return errors.Wrap(err, "error starting matreshka container")
	}

	err = b.dockerAPI.NetworkConnect(ctx, env.VervNetwork, cont.ID, &network.EndpointSettings{})
	if err != nil {
		return errors.Wrap(err, "error connecting matreshka container to verv network")
	}

	return nil
}

func (b *Matreshka) GetName() string {
	return containerName
}

func (b *Matreshka) GetDuration() time.Duration {
	return b.duration
}

func (b *Matreshka) IsAlive() (bool, error) {
	name := containerName

	containers, err := dockerutils.ListContainers(context.Background(), b.dockerAPI, &velez_api.ListSmerds_Request{
		Name: &name,
	})
	if err != nil {
		return false, errors.Wrap(err, "error listing smerds with name "+name)
	}

	for _, cont := range containers {
		hasName := false
		for _, cNname := range cont.Names {
			if name == cNname[1:] {
				hasName = true
				break
			}
		}

		if hasName && cont.State == velez_api.Smerd_running.String() {
			return true, nil
		}
	}

	return false, nil
}

func (b *Matreshka) Kill() error {
	err := b.dockerAPI.ContainerRemove(context.Background(), containerName, types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	})
	if err != nil {
		return errors.Wrap(err, "error dropping result")
	}

	return nil
}
