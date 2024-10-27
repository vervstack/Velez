package container_service_task

import (
	"context"
	"strings"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/godverv/Velez/internal/backservice/env"
	"github.com/godverv/Velez/internal/clients"
	"github.com/godverv/Velez/pkg/velez_api"
)

type Task[T any] struct {
	ApiClient       *ApiClient[T]
	Address         string
	grpcDialOpts    []grpc.DialOption
	grpcConstructor func(dial grpc.ClientConnInterface) T

	name            string
	containerConfig *container.Config

	hostConfig *container.HostConfig

	dockerAPI clients.Docker

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
		return errors.Wrap(err, "error creating container")
	}

	err = t.dockerAPI.ContainerStart(ctx, cont.ID, container.StartOptions{})
	if err != nil {
		return errors.Wrap(err, "error starting makosh container")
	}

	err = t.dockerAPI.NetworkConnect(ctx,
		env.VervNetwork,
		cont.ID,
		&network.EndpointSettings{Aliases: []string{t.name}})
	if err != nil {
		return errors.Wrap(err, "error connecting makosh container to verv network")
	}

	return nil
}

func (t *Task[T]) IsAlive() bool {
	ctx := context.Background()

	cont, err := t.dockerAPI.ContainerInspect(ctx, t.name)
	if err != nil {
		if strings.Contains(err.Error(), "No such container") {
			return false
		}
		logrus.Error(errors.Wrap(err, "error getting container of dependency: "+t.name))
		return false
	}

	if cont.State.Status != velez_api.Smerd_running.String() {
		return false
	}
	t.ApiClient, err = NewGrpcClient(t.Address, t.grpcConstructor, t.grpcDialOpts...)
	if err != nil {
		logrus.Error(errors.Wrap(err, "error creating grpc client for dependency in container: "+t.name))
		return false
	}

	if !t.healthCheck(t.ApiClient.Client) {
		return false
	}

	return true
}

func (t *Task[T]) Kill() error {
	ctx := context.Background()
	rmOpts := container.RemoveOptions{
		// TODO
		RemoveVolumes: true,
		Force:         true,
	}

	err := t.dockerAPI.ContainerRemove(ctx, t.name, rmOpts)
	if err != nil {
		if !strings.Contains(err.Error(), "No such container") {
			return errors.Wrap(err, "error dropping result")
		}
	}

	return nil
}

func (t *Task[T]) GetName() string {
	return t.name
}
