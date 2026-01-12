package verv_closed_network

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox/keep_alive"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients/headscale"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/cluster/env/container_service_task"
	"go.vervstack.ru/Velez/internal/domain"
	headscalePatterns "go.vervstack.ru/Velez/internal/patterns/headscale"
)

const (
	Name = "headscale"

	groupName         = "verv_private_network"
	defaultConfigPath = "/etc/headscale/config.yaml"
)

type headscaleLauncher struct {
	ctx context.Context

	clients node_clients.NodeClients
}

func SetupVcn(
	ctx context.Context, nodeClients node_clients.NodeClients) (
	cluster_clients.VervClosedNetworkClient, error) {
	state := nodeClients.LocalStateManager().Get()

	if !state.IsHeadscaleEnabled {
		// TODO return docker network wrapper
		return DisabledVcnImpl{}, nil
	}

	if state.HeadscaleKey != "" && state.HeadscaleServerUrl != "" {
		client, err := headscale.Connect(ctx, state.HeadscaleServerUrl, state.HeadscaleKey)
		if err != nil {
			return nil, rerrors.Wrap(err, "error creating headscale client")
		}

		return client, nil
	}

	client, err := LaunchHeadscale(ctx, nodeClients)
	if err != nil {
		return nil, rerrors.Wrap(err, "error launching headscale in this node")
	}

	return client, nil
}

func LaunchHeadscale(
	ctx context.Context,
	nodeClients node_clients.NodeClients,
) (*headscale.Client, error) {
	l := headscaleLauncher{ctx, nodeClients}

	err := l.deploy()
	if err != nil {
		return nil, rerrors.Wrap(err, "error starting headscale container")
	}

	client, err := headscale.ConnectToContainer(ctx, nodeClients, Name)
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating headscale client")
	}

	return client, nil
}

func (l headscaleLauncher) deploy() error {
	createContainerReq := headscalePatterns.Headscale(domain.SetupHeadscaleRequest{})

	taskConstructor, err := container_service_task.NewTaskV2(l.clients.Docker(), createContainerReq)
	if err != nil {
		return rerrors.Wrap(err, "error building task for headscale")
	}

	logrus.Info("Preparing HeadScale background task")

	// Launch
	_ = keep_alive.KeepAlive(
		taskConstructor,
		keep_alive.WithCancel(l.ctx.Done()),
		keep_alive.WithCheckInterval(time.Second/2),
	)

	return nil
}
