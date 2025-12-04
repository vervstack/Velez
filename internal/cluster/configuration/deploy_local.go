package configuration

import (
	"strconv"
	"strings"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/sirupsen/logrus"
	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"
	"go.redsock.ru/toolbox/closer"
	"go.redsock.ru/toolbox/keep_alive"
	version "go.vervstack.ru/matreshka/config"
	"golang.org/x/net/context"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients/matreshka"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/cluster/env/container_service_task"
	"go.vervstack.ru/Velez/internal/config"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	Name         = "matreshka"
	defaultImage = "vervstack/matreshka"
	grpcPort     = "50049"

	passEnv = "pass"

	defaultDataPath = "/app/data"
)

var image string

func init() {
	image = defaultImage + ":" + version.GetVersion()
}

func deployOnThisNode(ctx context.Context, cfg config.Config, nodeClients node_clients.NodeClients) (matreshka.Client, error) {
	key, err := initKey(ctx, nodeClients)
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting key")
	}

	matreshkaContainerCfg := container.CreateRequest{
		Config: &container.Config{
			Hostname: Name,
			Env: []string{
				passEnv + ":" + key,
			},
			ExposedPorts: map[nat.Port]struct{}{},
			Image:        toolbox.Coalesce(cfg.Environment.MatreshkaImage, image),
			Volumes: map[string]struct{}{
				Name: {},
			},
			Labels: map[string]string{
				labels.VervServiceLabel:  "true",
				labels.ComposeGroupLabel: Name,
			},
		},
		HostConfig: &container.HostConfig{
			PortBindings: map[nat.Port][]nat.PortBinding{
				grpcPort: {
					{
						HostPort: grpcPort,
					},
				},
			},
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeVolume,
					Source: Name,
					Target: defaultDataPath,
				},
			},
		},
	}

	if cfg.Environment.MatreshkaPort > 0 {
		matreshkaContainerCfg.Config.ExposedPorts[grpcPort] = struct{}{}

		matreshkaContainerCfg.HostConfig.PortBindings = map[nat.Port][]nat.PortBinding{
			grpcPort: {
				{
					HostPort: strconv.Itoa(cfg.Environment.MatreshkaPort),
				},
			},
		}
	}

	task, err := container_service_task.NewTaskV2(nodeClients.Docker(), matreshkaContainerCfg)
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating verv service task")
	}

	logrus.Info("Preparing matreshka service background task")

	ka := keep_alive.KeepAlive(task, keep_alive.WithCancel(ctx.Done()))

	if cfg.Environment.ShutDownOnExit {
		closer.Add(func() error {
			ka.Stop()
			return nil
		})
	}

	return nil, nil
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
