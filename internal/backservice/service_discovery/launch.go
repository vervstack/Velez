package service_discovery

import (
	"context"
	"strconv"
	"sync"
	"time"

	version "github.com/godverv/makosh/config"
	pb "github.com/godverv/makosh/pkg/makosh_be"
	"github.com/sirupsen/logrus"
	errors "go.redsock.ru/rerrors"
	rtb "go.redsock.ru/toolbox"
	"go.redsock.ru/toolbox/closer"
	"go.redsock.ru/toolbox/keep_alive"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/godverv/Velez/internal/backservice/env/container_service_task"
	"github.com/godverv/Velez/internal/clients"
	"github.com/godverv/Velez/internal/clients/makosh"
	"github.com/godverv/Velez/internal/config"
	"github.com/godverv/Velez/internal/security"
)

const (
	Name                 = "makosh"
	defaultImageBase     = "godverv/makosh"
	authTokenEnvVariable = "MAKOSH_ENVIRONMENT_AUTH-TOKEN"
	grpcPort             = "8080"
)

var image string

func init() {
	image = defaultImageBase + ":" + version.GetVersion()
}

var initModeSync = sync.Once{}

func LaunchMakosh(
	ctx context.Context,
	cfg *config.Config,
	clients clients.NodeClients,
) {
	initModeSync.Do(func() {
		var err error
		err = launchMakosh(ctx, cfg, clients)
		if err != nil {
			logrus.Fatal(errors.Wrap(err))
		}
	})
}

func launchMakosh(
	ctx context.Context,
	cfg *config.Config,
	nodeClients clients.NodeClients,
) error {
	// Construct
	token := string(rtb.RandomBase64(256))

	taskConstructor := container_service_task.NewTaskRequest[pb.MakoshBeAPIClient]{
		ContainerName:     Name,
		NodeClients:       nodeClients,
		ClientConstructor: pb.NewMakoshBeAPIClient,
		DialOpts:          nil,
		ImageName:         rtb.Coalesce(cfg.Environment.MakoshImage, image),
		GrpcPort:          grpcPort,
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

	taskConstructor.DialOpts = []grpc.DialOption{
		grpc.WithUnaryInterceptor(
			security.HeaderOutgoingInterceptor(makosh.AuthHeader, token)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
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
				Addrs:       []string{makoshTask.Address},
			},
		},
	}
	_, err = makoshTask.ApiClient.Client.UpsertEndpoints(ctx, req)
	if err != nil {
		logrus.Fatal(errors.Wrap(err, "error upserting makosh endpoint"))
	}

	// Change values in original config
	cfg.Environment.MakoshURL = makoshTask.Address
	cfg.Environment.MakoshKey = token

	return nil
}
