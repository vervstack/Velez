package container_service_task

import (
	"context"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/nat"
	"github.com/sirupsen/logrus"
	errors "go.redsock.ru/rerrors"
	"google.golang.org/grpc"

	"go.vervstack.ru/Velez/internal/backservice/env"
	"go.vervstack.ru/Velez/internal/clients"
	"go.vervstack.ru/Velez/internal/clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

type NewTaskRequest[T any] struct {
	NodeClients clients.NodeClients

	ClientConstructor func(dial grpc.ClientConnInterface) T
	DialOpts          []grpc.DialOption

	ContainerName string
	ImageName     string

	GrpcPort     string
	ExposedPorts map[string]string //container->host

	Healthcheck  func(client T) bool
	Env          map[string]string
	VolumeMounts map[string][]string // volume-name->container-file-path
}

func NewTask[T any](req NewTaskRequest[T]) (*Task[T], error) {
	if req.ClientConstructor == nil {
		return nil, errors.New("must provide client constructor")
	}

	if req.Healthcheck == nil {
		return nil, errors.New("must provide client healthcheck")
	}

	if req.GrpcPort == "" {
		return nil, errors.New("must provide grpc port to connect to")
	}

	dockerAPI := req.NodeClients.Docker().Client()
	portManager := req.NodeClients.PortManager()
	t := &Task[T]{
		grpcDialOpts:    req.DialOpts,
		grpcConstructor: req.ClientConstructor,
		healthCheck:     req.Healthcheck,

		name:            req.ContainerName,
		containerConfig: &container.Config{},
		hostConfig: &container.HostConfig{
			RestartPolicy: container.RestartPolicy{
				Name: container.RestartPolicyUnlessStopped,
			},
		},
		dockerAPI: dockerAPI,
	}
	ctx := context.Background()

	listImageReq := domain.ImageListRequest{Name: req.ImageName}
	images, err := dockerutils.ListImages(ctx, dockerAPI, listImageReq)
	if err != nil {
		return nil, errors.Wrap(err, "error listing images")
	}

	if len(images) == 0 {
		logrus.Infof("Image %s not found locally. Pulling it.", req.ImageName)
		_, err = dockerutils.PullImage(ctx, dockerAPI, req.ImageName, false)
		if err != nil {
			return nil, errors.Wrapf(err, "error pulling image %s", req.ImageName)
		}

		logrus.Infof("Image %s pulled successfully", req.ImageName)
	}

	// Container configuration
	t.containerConfig.Hostname = req.ContainerName
	t.containerConfig.Image = req.ImageName

	t.containerConfig.Labels = make(map[string]string)
	t.containerConfig.Labels[labels.CreatedWithVelezLabel] = "true"

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

		t.hostConfig.PortBindings[nat.Port(appendTCP(containerP))] = []nat.PortBinding{
			{
				HostPort: hostP,
			},
		}
	}
	if err != nil {
		portManager.UnlockPorts(occupiedPorts)
		return nil, errors.Wrap(err, "error getting ports")
	}

	for volume, paths := range req.VolumeMounts {
		for _, p := range paths {
			m := mount.Mount{
				Type:   mount.TypeVolume,
				Source: volume,
				Target: p,
			}

			t.hostConfig.Mounts = append(t.hostConfig.Mounts, m)
		}

	}

	if env.IsInContainer() {
		t.Address = req.ContainerName + ":" + req.GrpcPort
	} else {
		bindings := t.hostConfig.PortBindings[nat.Port(appendTCP(req.GrpcPort))]
		t.Address = "0.0.0.0:" + bindings[0].HostPort
	}

	logrus.Infof("Image %s running on address %s. Container name is: %s", req.ImageName, t.Address, t.name)

	return t, nil
}

func appendTCP(port string) string {
	if !strings.HasSuffix(port, "/tcp") {
		return port + "/tcp"
	}
	return port
}
