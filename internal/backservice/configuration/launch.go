package configuration

import (
	"strconv"
	"sync"

	"github.com/godverv/makosh/pkg/makosh_be"
	"github.com/sirupsen/logrus"
	errors "go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"
	"go.redsock.ru/toolbox/closer"
	"go.redsock.ru/toolbox/keep_alive"
	version "go.vervstack.ru/matreshka-be/config"
	"go.vervstack.ru/matreshka-be/pkg/matreshka_be_api"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/godverv/Velez/internal/backservice/env/container_service_task"
	"github.com/godverv/Velez/internal/backservice/service_discovery"
	"github.com/godverv/Velez/internal/clients"
	"github.com/godverv/Velez/internal/clients/matreshka"
	"github.com/godverv/Velez/internal/config"
)

const (
	Name         = "matreshka"
	defaultImage = "godverv/matreshka-be"
	grpcPort     = "80"
)

var image string

func init() {
	image = defaultImage + ":" + version.GetVersion()
}

var initOnce sync.Once

func LaunchMatreshka(ctx context.Context,
	cfg config.Config,
	clients clients.NodeClients,
	sd service_discovery.ServiceDiscovery,
) {
	initOnce.Do(func() {
		err := initInstance(ctx, cfg, clients, sd)
		if err != nil {
			logrus.Fatal(err)
		}
	})

	return
}

func initInstance(
	ctx context.Context,
	cfg config.Config,
	nodeClients clients.NodeClients,
	sd service_discovery.ServiceDiscovery,
) error {
	taskRequest := container_service_task.NewTaskRequest[matreshka_be_api.MatreshkaBeAPIClient]{
		NodeClients:       nodeClients,
		ClientConstructor: matreshka_be_api.NewMatreshkaBeAPIClient,
		DialOpts:          []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
		ContainerName:     Name,
		ImageName:         toolbox.Coalesce(cfg.Environment.MatreshkaImage, image),
		GrpcPort:          grpcPort,
		ExposedPorts:      map[string]string{},
		Healthcheck: func(client matreshka_be_api.MatreshkaBeAPIClient) bool {
			resp, err := client.ApiVersion(ctx, &matreshka_be_api.ApiVersion_Request{})
			if err == nil && resp != nil {
				return true
			}

			return false
		},
	}

	if cfg.Environment.MatreshkaPort > 0 {
		taskRequest.ExposedPorts[grpcPort] = strconv.Itoa(cfg.Environment.MatreshkaPort)
	} else {
		taskRequest.ExposedPorts[grpcPort] = ""
	}

	logrus.Info("Preparing matreshka service background task")

	task, err := container_service_task.NewTask(taskRequest)
	if err != nil {
		return errors.Wrap(err, "error creating task for matreshka")
	}

	ka := keep_alive.KeepAlive(task, keep_alive.WithCancel(ctx.Done()))
	if cfg.Environment.ShutDownOnExit {
		closer.Add(func() error {
			ka.Stop()
			return nil
		})
	}

	matreshkaEndpoints := &makosh_be.UpsertEndpoints_Request{
		Endpoints: []*makosh_be.Endpoint{
			{
				ServiceName: matreshka.ServiceName,
				Addrs:       []string{task.Address},
			},
		},
	}

	_, err = sd.UpsertEndpoints(ctx, matreshkaEndpoints)
	if err != nil {
		return errors.Wrap(err, "error upserting endpoints for matreshka to makosh")
	}

	return nil
}
