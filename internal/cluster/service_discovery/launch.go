package service_discovery

import (
	"context"
	"strconv"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"go.redsock.ru/rerrors"
	rtb "go.redsock.ru/toolbox"
	"go.redsock.ru/toolbox/closer"
	"go.redsock.ru/toolbox/keep_alive"
	version "go.vervstack.ru/makosh/config"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients/makosh"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/cluster/env"
	"go.vervstack.ru/Velez/internal/cluster/env/container_service_task"
	"go.vervstack.ru/Velez/internal/config"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/pipelines"
	"go.vervstack.ru/Velez/internal/pipelines/steps"
)

const (
	Name                 = "makosh"
	defaultImageBase     = "vervstack/makosh"
	authTokenEnvVariable = "MAKOSH_ENVIRONMENT_AUTH-TOKEN"
	grpcPort             = "8080/tcp"
)

var image string

func init() {
	image = defaultImageBase + ":" + version.GetVersion()
}

func SetupMakosh(
	ctx context.Context,
	cfg config.Config,
	nodeClients node_clients.NodeClients,
	vcnClient cluster_clients.VervClosedNetworkClient,
) (sd *makosh.ServiceDiscovery, err error) {
	// TODO statefull token?
	token := string(rtb.RandomBase64(256))

	containerReq := container.CreateRequest{
		Config: &container.Config{
			Hostname: Name,
			Image:    rtb.Coalesce(cfg.Environment.MakoshImage, image),
			Env: []string{
				authTokenEnvVariable + ":" + token,
			},
			ExposedPorts: map[nat.Port]struct{}{},
			Labels: map[string]string{
				labels.ComposeGroupLabel: Name,
			},
		},
		HostConfig: &container.HostConfig{},
	}

	if cfg.Environment.MakoshPort > 0 || !env.IsInContainer() {
		port := strconv.Itoa(cfg.Environment.MakoshPort)
		if cfg.Environment.MakoshPort == 0 {
			port = ""
		}
		strconv.Itoa(cfg.Environment.MakoshPort)
		containerReq.ExposedPorts[grpcPort] = struct{}{}
		containerReq.HostConfig.PortBindings = nat.PortMap{
			grpcPort: []nat.PortBinding{
				{
					HostPort: port,
				},
			},
		}
	}

	task, err := container_service_task.NewTaskV2(nodeClients.Docker(), containerReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating makosh service task")
	}

	// Launch
	keepAlive := keep_alive.KeepAlive(
		task,
		keep_alive.WithCancel(ctx.Done()),
		keep_alive.WithCheckInterval(time.Second/2),
	)

	if cfg.Environment.ShutDownOnExit {
		closer.Add(func() error {
			keepAlive.Stop()
			return nil
		})
	}

	if !env.IsInContainer() {
		// When running outside of container (e.g. direct binary execution / local debug)
		// changing target address to accessible from localhost
		addr, port := task.GetPortBinding(grpcPort)
		cfg.Environment.MakoshURL = addr + ":" + port
	}

	sd, err = makosh.NewServiceDiscovery(cfg)
	if err != nil {
		return sd, rerrors.Wrap(err, "error initializing service discovery ")
	}

	connToVpnReq := domain.ConnectServiceToVcn{
		ServiceName: Name,
	}

	runner := pipelines.ConnectServiceToVpn(connToVpnReq, nodeClients, vcnClient, sd)
	err = runner.Run(ctx)
	if err != nil {
		if !rerrors.Is(err, steps.ErrAlreadyExists) {
			return sd, rerrors.Wrap(err, "error connecting service to verv closed network")
		}
	}

	return sd, nil
}
