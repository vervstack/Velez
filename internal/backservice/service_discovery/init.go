package service_discovery

import (
	"context"
	"strconv"
	"sync"
	"time"

	rtb "github.com/Red-Sock/toolbox"
	"github.com/Red-Sock/toolbox/keep_alive"
	errors "github.com/Red-Sock/trace-errors"
	pb "github.com/godverv/makosh/pkg/makosh_be"
	"github.com/sirupsen/logrus"

	"github.com/godverv/Velez/internal/backservice/env/container_service_task"
	"github.com/godverv/Velez/internal/clients"
	"github.com/godverv/Velez/internal/clients/makosh"
	"github.com/godverv/Velez/internal/config"
)

var initModeSync = sync.Once{}

func InitInstance(
	ctx context.Context,
	cfg *config.Config,
	clients clients.NodeClients,
) {
	initModeSync.Do(func() {
		var err error
		err = launchServiceDiscovery(ctx, cfg, clients)
		if err != nil {
			logrus.Fatal(errors.Wrap(err))
		}
	})
}

func launchServiceDiscovery(
	ctx context.Context,
	cfg *config.Config,
	nodeClients clients.NodeClients,
) error {
	// Construct
	token := string(rtb.RandomBase64(256))

	var taskConstructor container_service_task.NewTaskRequest[pb.MakoshBeAPIClient]

	taskConstructor.NodeClients = nodeClients

	taskConstructor.ContainerName = Name
	taskConstructor.ImageName = rtb.Coalesce(cfg.Environment.MakoshImageName, image)

	taskConstructor.ExposedPorts = map[string]string{}
	taskConstructor.Env = map[string]string{
		makoshContainerAuthTokenEnvVariable: token,
	}

	if cfg.Environment.MakoshExposePort {
		taskConstructor.ExposedPorts[strconv.Itoa(cfg.Environment.MakoshPort)] = ""
	}

	taskConstructor.ClientConstructor = pb.NewMakoshBeAPIClient
	makoshTask, err := container_service_task.NewTask[pb.MakoshBeAPIClient](taskConstructor)
	if err != nil {
		return errors.Wrap(err, "error creating task")
	}
	// Launch
	logrus.Info("Starting service discovery background task")
	keepAlive := keep_alive.KeepAlive(
		makoshTask,
		keep_alive.WithCancel(ctx.Done()),
		keep_alive.WithCheckInterval(time.Second/2),
	)
	keepAlive.Wait()

	// Add self to makosh
	req := &pb.UpsertEndpoints_Request{
		Endpoints: []*pb.Endpoint{
			{
				ServiceName: makosh.ServiceName,
				Addrs:       []string{makoshTask.Address},
			},
		},
	}
	_, err = makoshTask.ApiClient.Client.UpsertEndpoints(ctx, req)
	if err != nil {
		logrus.Fatal(errors.Wrap(err, "error upserting makosh endpoint"))
	}

	// Change values in original config
	cfg.Environment.MakoshUrl = makoshTask.Address
	cfg.Environment.MakoshKey = token

	return nil
}
