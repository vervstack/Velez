package configuration

import (
	"strconv"
	"sync"

	"github.com/sirupsen/logrus"
	errors "go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox/keep_alive"
	"go.verv.tech/matreshka-be/pkg/matreshka_be_api"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/godverv/Velez/internal/backservice/env/container_service_task"
	"github.com/godverv/Velez/internal/clients"
	"github.com/godverv/Velez/internal/config"
)

const (
	Name  = "matreshka"
	image = "godverv/matreshka-be:v1.0.41"
)

var initOnce sync.Once

type MatreshkaConnect struct {
	Addr      string
	AuthToken string
}

var localMatreshka MatreshkaConnect

func LaunchMatreshka(ctx context.Context, cfg config.Config, clients clients.NodeClients) MatreshkaConnect {
	initOnce.Do(func() {
		err := initInstance(ctx, cfg, clients)
		if err != nil {
			logrus.Fatal(err)
		}
	})

	return localMatreshka
}

func initInstance(
	ctx context.Context,
	cfg config.Config,
	nodeClients clients.NodeClients,
) error {
	taskRequest := container_service_task.NewTaskRequest[matreshka_be_api.MatreshkaBeAPIClient]{
		NodeClients:       nodeClients,
		ClientConstructor: matreshka_be_api.NewMatreshkaBeAPIClient,
		DialOpts:          []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
		ContainerName:     Name,
		ImageName:         image,
		GrpcPort:          "80",
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
		taskRequest.ExposedPorts["80"] = strconv.Itoa(cfg.Environment.MatreshkaPort)
	} else {
		taskRequest.ExposedPorts["80"] = ""
	}

	task, err := container_service_task.NewTask(taskRequest)
	if err != nil {
		return errors.Wrap(err, "error creating task for matreshka")
	}

	logrus.Info("Starting matreshka service background task")
	ka := keep_alive.KeepAlive(task, keep_alive.WithCancel(ctx.Done()))
	ka.Wait()

	localMatreshka.Addr = task.Address

	return nil
}
