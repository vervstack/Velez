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

	"go.vervstack.ru/Velez/internal/backservice/env"
	"go.vervstack.ru/Velez/internal/clients"
	"go.vervstack.ru/Velez/internal/clients/docker"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

type Task[T any] struct {
	ApiClient            *ApiClient[T]
	ContainerNetworkHost string

	createClient func(t *Task[T]) (*ApiClient[T], error)

	name            string
	containerConfig *container.Config

	hostConfig *container.HostConfig

	docker    clients.Docker
	dockerAPI client.APIClient

	healthCheck func(client T) bool
}

func (t *Task[T]) Start() error {
	ctx := context.Background()
	cont, err := t.dockerAPI.ContainerCreate(ctx,
		t.containerConfig,
		t.hostConfig,
		&network.NetworkingConfig{},
		&v1.Platform{},
		t.name,
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
		&network.EndpointSettings{Aliases: []string{t.name}})
	if err != nil {
		return rerrors.Wrap(err, "error connecting makosh container to verv network")
	}

	return nil
}

func (t *Task[T]) IsAlive() bool {
	ctx := context.Background()

	cont, err := t.dockerAPI.ContainerInspect(ctx, t.name)
	if err != nil {
		if strings.Contains(err.Error(), docker.NoSuchContainerError) {
			return false
		}
		logrus.Error(rerrors.Wrap(err, "error getting container of dependency: "+t.name))
		return false
	}

	if cont.State.Status != velez_api.Smerd_running.String() {
		return false
	}

	if cont.Config.Image != t.containerConfig.Image {
		return false
	}

	if t.createClient != nil && t.ApiClient == nil {
		t.ApiClient, err = t.createClient(t)
		if err != nil {
			logrus.Error(rerrors.Wrap(err, "error creating grpc client for dependency in container: "+t.name))
			return false
		}
	}

	if t.ApiClient == nil || t.healthCheck == nil {
		return true
	}

	if !t.healthCheck(t.ApiClient.Client) {
		return false
	}

	return true
}

func (t *Task[T]) Kill() error {
	ctx := context.Background()

	err := t.docker.Remove(ctx, t.name)
	if err != nil {
		if !strings.Contains(err.Error(), docker.NoSuchContainerError) {
			return rerrors.Wrap(err, "error dropping result")
		}
	}

	return nil
}

func (t *Task[T]) GetName() string {
	return t.name
}

func (t *Task[T]) GetPortBinding(port string) string {
	if env.IsInContainer() {
		return port
	}

	bindings := t.hostConfig.PortBindings[nat.Port(appendTCP(port))]
	return bindings[0].HostPort
}
