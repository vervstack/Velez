package configuration

import (
	"context"
	"strconv"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/nat"
	"github.com/sirupsen/logrus"
	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"
	"go.redsock.ru/toolbox/closer"
	"go.redsock.ru/toolbox/keep_alive"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients/makosh"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients/matreshka"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/cluster/env/container_service_task"
	"go.vervstack.ru/Velez/internal/config"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/pipelines"
	"go.vervstack.ru/Velez/internal/pipelines/steps"
)

func SetupMatreshka(
	ctx context.Context,
	cfg config.Config,
	nc node_clients.NodeClients,
	sdClient *makosh.ServiceDiscovery,
	vcnClient cluster_clients.VervClosedNetworkClient,
) (matreshka.Client, error) {
	key, err := initKey(ctx, nc)
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting key")
	}

	//region Create Container request
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
	//endregion

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

	task, err := container_service_task.NewTaskV2(nc.Docker(), matreshkaContainerCfg)
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

	mClient, err := newClient(nc)
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating matreshka client")
	}

	vcnReq := domain.ConnectServiceToVcn{
		ServiceName: Name,
	}

	runner := pipelines.ConnectServiceToVpn(vcnReq, nc, vcnClient, sdClient)
	err = runner.Run(ctx)
	if err != nil {
		if rerrors.Is(err, steps.ErrAlreadyExists) {
			return mClient, nil
		}

		return nil, rerrors.Wrap(err, "error connecting matreshka to verv closed network")
	}

	return mClient, nil
}
