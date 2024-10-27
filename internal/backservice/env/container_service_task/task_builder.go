package container_service_task

import (
	"context"
	"strconv"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"google.golang.org/grpc"

	"github.com/godverv/Velez/internal/clients"
	"github.com/godverv/Velez/internal/clients/docker/dockerutils"
	"github.com/godverv/Velez/internal/service/service_manager/container_manager_v1"
)

type NewTaskRequest[T any] struct {
	NodeClients clients.NodeClients

	ClientConstructor func(dial grpc.ClientConnInterface) T

	ContainerName string
	ImageName     string
	ExposedPorts  map[string]string //container->host
	Healthcheck   func(client T) bool
	Env           map[string]string
}

func NewTask[T any](req NewTaskRequest[T]) (*Task[T], error) {
	dockerAPI := req.NodeClients.Docker()
	portManager := req.NodeClients.PortManager()
	t := &Task[T]{
		name:            req.ContainerName,
		containerConfig: &container.Config{},
		hostConfig:      &container.HostConfig{},
		dockerAPI:       dockerAPI,
	}
	ctx := context.Background()

	_, err := dockerutils.PullImage(ctx, dockerAPI, req.ImageName, false)
	if err != nil {
		return nil, errors.Wrap(err, "error pulling image")
	}

	// Container configuration
	t.containerConfig.Hostname = req.ContainerName
	t.containerConfig.Image = req.ImageName

	t.containerConfig.Labels = make(map[string]string)
	t.containerConfig.Labels[container_manager_v1.CreatedWithVelezLabel] = "true"

	t.containerConfig.Env = make([]string, 0, len(t.containerConfig.Env))
	for k, v := range req.Env {
		t.containerConfig.Env = append(t.containerConfig.Env, k+"="+v)
	}

	occupiedPorts := make([]uint32, 0, len(req.ExposedPorts))

	// Host configuration
	t.hostConfig.PortBindings = make(nat.PortMap, len(req.ExposedPorts))
	for containerP, hostP := range req.ExposedPorts {
		if hostP == "" {
			var p uint32
			p, err = portManager.GetPort()
			if err != nil {
				break
			}

			occupiedPorts = append(occupiedPorts, p)
			hostP = strconv.FormatUint(uint64(p), 10)
		}

		t.hostConfig.PortBindings[nat.Port(containerP)] = []nat.PortBinding{
			{
				HostPort: hostP,
			},
		}
	}

	if err != nil {
		portManager.UnlockPorts(occupiedPorts)
		return nil, errors.Wrap(err, "error getting ports")
	}

	return t, nil
}
