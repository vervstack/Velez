package configuration

import (
	"context"
	"time"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/godverv/Velez/internal/backservice/env"
	"github.com/godverv/Velez/internal/clients/docker/dockerutils"
	"github.com/godverv/Velez/pkg/velez_api"
)

const (
	Name     = "matreshka"
	image    = "godverv/matreshka-be:v1.0.23"
	duration = time.Second * 5
)

type Matreshka struct {
	dockerAPI client.CommonAPIClient

	duration time.Duration

	port string
}

func New(dockerAPI client.CommonAPIClient, exposeToPort string) *Matreshka {
	w := &Matreshka{
		dockerAPI: dockerAPI,
		duration:  duration,
		port:      exposeToPort,
	}

	return w
}

func (b *Matreshka) Start() error {
	isAlive, err := b.IsAlive()
	if err != nil || isAlive {
		return err
	}

	ctx := context.Background()

	_, err = dockerutils.PullImage(ctx, b.dockerAPI, image, false)
	if err != nil {
		return errors.Wrap(err, "error pulling matreshka image")
	}

	hostConf := &container.HostConfig{}

	if b.port != "" {
		hostConf.PortBindings = nat.PortMap{
			"80/tcp": []nat.PortBinding{
				{
					HostPort: b.port,
				},
			},
		}
	}

	cont, err := b.dockerAPI.ContainerCreate(ctx,
		&container.Config{
			Image:    image,
			Hostname: Name,
		},
		hostConf,
		&network.NetworkingConfig{},
		&v1.Platform{},
		Name,
	)
	if err != nil {
		return errors.Wrap(err, "error creating matreshka container")
	}

	err = b.dockerAPI.ContainerStart(ctx, cont.ID, types.ContainerStartOptions{})
	if err != nil {
		return errors.Wrap(err, "error starting matreshka container")
	}

	err = b.dockerAPI.NetworkConnect(ctx, env.VervNetwork, cont.ID, &network.EndpointSettings{
		Aliases: []string{Name},
	})
	if err != nil {
		return errors.Wrap(err, "error connecting matreshka container to verv network")
	}

	return nil
}

func (b *Matreshka) GetName() string {
	return Name
}

func (b *Matreshka) GetDuration() time.Duration {
	return b.duration
}

func (b *Matreshka) IsAlive() (bool, error) {
	name := Name

	containers, err := dockerutils.ListContainers(
		context.Background(),
		b.dockerAPI, &velez_api.ListSmerds_Request{
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
	err := b.dockerAPI.ContainerRemove(context.Background(), Name, types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	})
	if err != nil {
		return errors.Wrap(err, "error dropping result")
	}

	return nil
}
