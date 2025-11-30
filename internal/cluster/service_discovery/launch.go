package service_discovery

import (
	"context"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"go.redsock.ru/rerrors"
	rtb "go.redsock.ru/toolbox"
	"go.redsock.ru/toolbox/closer"
	"go.redsock.ru/toolbox/keep_alive"
	version "go.vervstack.ru/makosh/config"
	pb "go.vervstack.ru/makosh/pkg/makosh_be"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients/makosh"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/cluster/env/container_service_task"
	"go.vervstack.ru/Velez/internal/config"
	"go.vervstack.ru/Velez/internal/middleware"
)

const (
	Name                 = "makosh"
	defaultImageBase     = "vervstack/makosh"
	authTokenEnvVariable = "MAKOSH_ENVIRONMENT_AUTH-TOKEN"
	grpcPort             = "8080"
)

var image string

func init() {
	image = defaultImageBase + ":" + version.GetVersion()
}

func SetupMakosh(
	ctx context.Context,
	cfg config.Config,
	nodeClients node_clients.NodeClients,
) (sd makosh.ServiceDiscovery, err error) {
	token := string(rtb.RandomBase64(256))

	taskConstructor := container_service_task.NewTaskRequest[pb.MakoshBeAPIClient]{
		ContainerName: Name,
		NodeClients:   nodeClients,

		ImageName: rtb.Coalesce(cfg.Environment.MakoshImage, image),
		ExposedPorts: map[string]string{
			grpcPort: "",
		},
		Env: map[string]string{
			authTokenEnvVariable: token,
		},
	}

	if cfg.Environment.MakoshPort > 0 {
		taskConstructor.ExposedPorts[grpcPort] = strconv.Itoa(cfg.Environment.MakoshPort)
	}

	var taskClient *container_service_task.ApiClient[pb.MakoshBeAPIClient]

	taskConstructor.Healthcheck = func(t *container_service_task.Task[pb.MakoshBeAPIClient]) bool {
		if taskClient == nil {
			taskClient, err = initClient(t, token)
			if err != nil {
				return false
			}
		}

		resp, err := taskClient.Client.Version(ctx, &pb.Version_Request{})
		if err != nil && resp == nil {
			return false
		}
		return true
	}

	logrus.Info("Preparing service discovery background task")

	makoshTask, err := container_service_task.NewTask[pb.MakoshBeAPIClient](taskConstructor)
	if err != nil {
		return sd, rerrors.Wrap(err, "error creating task")
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

	req := &pb.UpsertEndpoints_Request{
		Endpoints: []*pb.Endpoint{
			{
				ServiceName: makosh.ServiceName,
				Addrs:       []string{taskClient.Addr},
			},
		},
	}

	_, err = taskClient.Client.UpsertEndpoints(ctx, req)
	if err != nil {
		return sd, rerrors.Wrap(err, "error upserting makosh endpoint")
	}

	cfg.Environment.MakoshURL = taskClient.Addr
	cfg.Environment.MakoshKey = token

	sd, err = makosh.NewServiceDiscovery(cfg)
	if err != nil {
		return sd, rerrors.Wrap(err, "error initializing service discovery ")
	}

	return sd, nil
}

func initClient(t *container_service_task.Task[pb.MakoshBeAPIClient], token string) (
	*container_service_task.ApiClient[pb.MakoshBeAPIClient], error) {
	return container_service_task.NewGrpcClient(
		t.ContainerNetworkHost+":"+t.GetPortBinding(grpcPort),
		pb.NewMakoshBeAPIClient,
		grpc.WithChainUnaryInterceptor(middleware.HeaderOutgoingInterceptor(makosh.AuthHeader, token)),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
}
