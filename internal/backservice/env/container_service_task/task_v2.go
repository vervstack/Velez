package container_service_task

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sirupsen/logrus"
	"go.redsock.ru/rerrors"
	rtb "go.redsock.ru/toolbox"

	"go.vervstack.ru/Velez/internal/backservice/env"
	"go.vervstack.ru/Velez/internal/clients"
	"go.vervstack.ru/Velez/internal/clients/docker"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

type TaskV2 struct {
	container container.CreateRequest

	docker    clients.Docker
	dockerAPI client.APIClient
}

func NewTaskV2(docker clients.Docker, ctr container.CreateRequest) (*TaskV2, error) {
	if ctr.Config == nil {
		return nil, rerrors.New("config is nil")
	}

	if ctr.Config.Hostname == "" {
		return nil, rerrors.New("hostname is empty")
	}

	ctr.HostConfig = rtb.Coalesce(ctr.HostConfig, &container.HostConfig{})
	ctr.NetworkingConfig = rtb.Coalesce(ctr.NetworkingConfig, &network.NetworkingConfig{})

	return &TaskV2{
		container: ctr,
		docker:    docker,
		dockerAPI: docker.Client(),
	}, nil
}

func (t *TaskV2) Start() error {
	ctx := context.Background()
	cont, err := t.dockerAPI.ContainerCreate(ctx,
		t.container.Config,
		t.container.HostConfig,
		t.container.NetworkingConfig,
		&v1.Platform{},
		t.container.Hostname,
	)
	if err != nil {
		return rerrors.Wrap(err, "error creating container")
	}

	err = t.dockerAPI.ContainerStart(ctx, cont.ID, container.StartOptions{})
	if err != nil {
		return rerrors.Wrap(err, "error starting makosh container")
	}

	err = t.dockerAPI.NetworkConnect(ctx,
		env.VervNetwork,
		cont.ID,
		&network.EndpointSettings{Aliases: []string{t.container.Hostname}})
	if err != nil {
		return rerrors.Wrap(err, "error connecting makosh container to verv network")
	}

	return nil
}

func (t *TaskV2) IsAlive() bool {
	ctx := context.Background()

	cont, err := t.dockerAPI.ContainerInspect(ctx, t.container.Hostname)
	if err != nil {
		if strings.Contains(err.Error(), docker.NoSuchContainerError) {
			return false
		}
		logrus.Error(rerrors.Wrap(err, "error getting container of dependency: "+t.container.Hostname))
		return false
	}

	if cont.State.Status != velez_api.Smerd_running.String() {
		return false
	}

	if cont.Config.Image != t.container.Image {
		return false
	}

	return true
}

func (t *TaskV2) Kill() error {
	ctx := context.Background()

	err := t.docker.Remove(ctx, t.container.Hostname)
	if err != nil {
		if !strings.Contains(err.Error(), docker.NoSuchContainerError) {
			return rerrors.Wrap(err, "error dropping result")
		}
	}

	return nil
}

func (t *TaskV2) GetName() string {
	return t.container.Hostname
}

func (t *TaskV2) GetPortBinding(port string) string {
	if env.IsInContainer() {
		return port
	}

	bindings := t.container.HostConfig.PortBindings[nat.Port(appendTCP(port))]
	return bindings[0].HostPort
}
