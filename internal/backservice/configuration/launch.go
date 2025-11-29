package configuration

import (
	"strconv"
	"strings"
	"sync"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"
	"go.redsock.ru/toolbox/closer"
	"go.redsock.ru/toolbox/keep_alive"
	version "go.vervstack.ru/matreshka/config"
	"go.vervstack.ru/matreshka/pkg/app/matreshka_client"
	"go.vervstack.ru/matreshka/pkg/matreshka_api"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.vervstack.ru/Velez/internal/backservice/env/container_service_task"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
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
	clients node_clients.NodeClients,
) {
	initOnce.Do(func() {
		err := initInstance(ctx, cfg, clients)
		if err != nil {
			logrus.Fatal(err)
		}
	})

	return
}

func initInstance(
	ctx context.Context,
	cfg *config.Config,
	nodeClients node_clients.NodeClients,
) (err error) {
	key, err := initKey(ctx, nodeClients)
	if err != nil {
		return rerrors.Wrap(err, "error getting key")
	}

	taskRequest := container_service_task.NewTaskRequest[matreshka_api.MatreshkaBeAPIClient]{
		NodeClients: nodeClients,

		ContainerName: Name,
		ImageName:     toolbox.Coalesce(cfg.Environment.MatreshkaImage, image),
		ExposedPorts:  map[string]string{},
		Env: map[string]string{
			passEnv: key,
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

	var apiClient *container_service_task.ApiClient[matreshka_api.MatreshkaBeAPIClient]

	taskRequest.Healthcheck = func(t *container_service_task.Task[matreshka_api.MatreshkaBeAPIClient]) bool {
		if apiClient == nil {
			apiClient, err = initClient(t, key)
			if err != nil {
				return false
			}
		}

		resp, err := apiClient.Client.ApiVersion(ctx, &matreshka_api.ApiVersion_Request{})
		if err == nil && resp != nil {
			return true
		}

		return false
	}

	task, err := container_service_task.NewTask(taskRequest)
	if err != nil {
		return rerrors.Wrap(err, "error creating task for matreshka")
	}

	ka := keep_alive.KeepAlive(task, keep_alive.WithCancel(ctx.Done()))

	if cfg.Environment.ShutDownOnExit {
		closer.Add(func() error {
			ka.Stop()
			return nil
		})
	}

	return nil
}

func initKey(ctx context.Context, nodeClients node_clients.NodeClients) (string, error) {
	keyFromLocalState := nodeClients.LocalStateManager().Get().MatreshkaKey

	keyFromCont, err := getKeyFromMatreshkaContainerEnv(ctx, nodeClients.Docker().Client())
	if err != nil {
		return "", rerrors.Wrap(err, "error getting key from container")
	}

	if keyFromCont == "" {
		return keyFromLocalState, nil
	}

	logrus.Infof("Using key from local state: %s", keyFromLocalState)

	stateManager := nodeClients.LocalStateManager()
	localState := stateManager.Get()
	localState.MatreshkaKey = keyFromCont
	stateManager.Set(localState)

	return keyFromCont, nil
}

func getKeyFromMatreshkaContainerEnv(ctx context.Context, docker client.APIClient) (string, error) {
	cont, err := docker.ContainerInspect(ctx, Name)
	if err != nil {
		if !cerrdefs.IsNotFound(err) {
			return "", rerrors.Wrap(err, "")
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

func initClient(t *container_service_task.Task[matreshka_api.MatreshkaBeAPIClient],
	key string) (*container_service_task.ApiClient[matreshka_api.MatreshkaBeAPIClient], error) {
	return container_service_task.NewGrpcClient(
		t.ContainerNetworkHost+":"+t.GetPortBinding(grpcPort),
		matreshka_api.NewMatreshkaBeAPIClient,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(matreshka_client.WithHeader(matreshka_client.Pass, key)))
}
