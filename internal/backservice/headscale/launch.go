package headscale

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	errors "go.redsock.ru/rerrors"
	rtb "go.redsock.ru/toolbox"
	"go.redsock.ru/toolbox/closer"
	"go.redsock.ru/toolbox/keep_alive"
	pb "go.vervstack.ru/makosh/pkg/makosh_be"

	"go.vervstack.ru/Velez/internal/backservice/env/container_service_task"
	"go.vervstack.ru/Velez/internal/clients"
	"go.vervstack.ru/Velez/internal/clients/makosh"
	"go.vervstack.ru/Velez/internal/config"
)

const (
	Name         = "headscale"
	defaultImage = "docker.io/headscale/headscale:v0"
)

var initModeSync = sync.Once{}

func Launch(
	ctx context.Context,
	cfg *config.Config,
	clients clients.NodeClients,
) {
	initModeSync.Do(func() {
		var err error
		err = launch(ctx, cfg, clients)
		if err != nil {
			logrus.Fatal(errors.Wrap(err))
		}
	})
}

func launch(
	ctx context.Context,
	cfg *config.Config,
	nodeClients clients.NodeClients,
) error {
	taskConstructor := container_service_task.NewTaskRequest[pb.MakoshBeAPIClient]{
		ContainerName: Name,
		NodeClients:   nodeClients,

		ImageName: rtb.Coalesce(cfg.Environment.HeadscaleImage, defaultImage),
		ExposedPorts: map[string]string{
			grpcPort: "",
		},
		Healthcheck: nil,
		Env: map[string]string{
			authTokenEnvVariable: token,
		},
	}

	if cfg.Environment.MakoshPort > 0 {
		taskConstructor.ExposedPorts[grpcPort] = strconv.Itoa(cfg.Environment.MakoshPort)
	}

	taskConstructor.Healthcheck = func(client pb.MakoshBeAPIClient) bool {
		resp, err := client.Version(ctx, &pb.Version_Request{})
		if err != nil {
			return false
		}

		if resp == nil {
			return false
		}

		return true
	}

	logrus.Info("Preparing service discovery background task")
	makoshTask, err := container_service_task.NewTask[pb.MakoshBeAPIClient](taskConstructor)
	if err != nil {
		return errors.Wrap(err, "error creating task")
	}
	// Launch
	keepAlive := keep_alive.KeepAlive(
		makoshTask,
		keep_alive.WithCancel(ctx.Done()),
		keep_alive.WithCheckInterval(time.Second/2),
	)
	if cfg.Environment.ShutDownOnExit {
		closer.Add(func() error {
			keepAlive.Stop()
			return nil
		})
	}

	// Add self to makosh
	req := &pb.UpsertEndpoints_Request{
		Endpoints: []*pb.Endpoint{
			{
				ServiceName: makosh.ServiceName,
				Addrs:       []string{makoshTask.ContainerNetworkHost},
			},
		},
	}
	_, err = makoshTask.ApiClient.Client.UpsertEndpoints(ctx, req)
	if err != nil {
		logrus.Fatal(errors.Wrap(err, "error upserting makosh endpoint"))
	}

	// Change values in original config
	cfg.Environment.MakoshURL = makoshTask.ContainerNetworkHost
	cfg.Environment.MakoshKey = token

	return nil
}
