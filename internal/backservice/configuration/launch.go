package configuration

import (
	"strconv"
	"strings"
	"sync"

	"github.com/docker/docker/errdefs"
	"github.com/sirupsen/logrus"
	errors "go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"
	"go.redsock.ru/toolbox/closer"
	"go.redsock.ru/toolbox/keep_alive"
	"go.vervstack.ru/makosh/pkg/makosh_be"
	version "go.vervstack.ru/matreshka/config"
	"go.vervstack.ru/matreshka/pkg/app/matreshka_client"
	"go.vervstack.ru/matreshka/pkg/matreshka_api"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.vervstack.ru/Velez/internal/backservice/env/container_service_task"
	"go.vervstack.ru/Velez/internal/backservice/service_discovery"
	"go.vervstack.ru/Velez/internal/clients"
	"go.vervstack.ru/Velez/internal/clients/matreshka"
	"go.vervstack.ru/Velez/internal/config"
)

const (
	Name         = "matreshka"
	defaultImage = "vervstack/matreshka"
	grpcPort     = "50049"

	passEnv = "pass"
)

var image string

func init() {
	image = defaultImage + ":" + version.GetVersion()
}

var initOnce sync.Once

func LaunchMatreshka(ctx context.Context,
	cfg *config.Config,
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
	cfg *config.Config,
	nodeClients clients.NodeClients,
	sd service_discovery.ServiceDiscovery,
) (err error) {
	if cfg.Environment.MatreshkaKey == "" {
		cfg.Environment.MatreshkaKey, err = getKeyFromContainer(ctx, nodeClients.Docker())
		if err != nil {
			return errors.Wrap(err)
		}
	}

	if cfg.Environment.MatreshkaKey == "" {
		generateKey(cfg)
		logrus.Infof("matreshka key not set. Generating one: %s", cfg.Environment.MatreshkaKey)
	} else {
		logrus.Infof("matreshka key is set to %s", cfg.Environment.MatreshkaKey)

	}

	taskRequest := container_service_task.NewTaskRequest[matreshka_api.MatreshkaBeAPIClient]{
		NodeClients:       nodeClients,
		ClientConstructor: matreshka_api.NewMatreshkaBeAPIClient,
		DialOpts: []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithUnaryInterceptor(matreshka_client.WithHeader(matreshka_client.Pass, cfg.Environment.MatreshkaKey))},
		ContainerName: Name,
		ImageName:     toolbox.Coalesce(cfg.Environment.MatreshkaImage, image),
		GrpcPort:      grpcPort,
		ExposedPorts:  map[string]string{},
		Healthcheck: func(client matreshka_api.MatreshkaBeAPIClient) bool {
			resp, err := client.ApiVersion(ctx, &matreshka_api.ApiVersion_Request{})
			if err == nil && resp != nil {
				return true
			}

			return false
		},
		Env: map[string]string{
			passEnv: cfg.Environment.MatreshkaKey,
		},
		VolumeMounts: map[string][]string{
			"matreshka": {"/app/data"},
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

func getKeyFromContainer(ctx context.Context, docker clients.Docker) (string, error) {
	cont, err := docker.InspectContainer(ctx, Name)
	if err != nil {
		if !errdefs.IsNotFound(err) {
			return "", errors.Wrap(err, "")
		}

		return "", nil
	}
	for _, e := range cont.Config.Env {
		if strings.HasPrefix(e, passEnv) {
			return e[len(passEnv)+1:], nil
		}
	}
	return "", nil
}
func generateKey(cfg *config.Config) {
	cfg.Environment.MatreshkaKey = string(toolbox.RandomBase64(256))
}
