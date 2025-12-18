package verv_closed_network

import (
	"context"
	"time"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/sirupsen/logrus"
	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"
	"go.redsock.ru/toolbox/keep_alive"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients/headscale"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/cluster/env/container_service_task"
	"go.vervstack.ru/Velez/internal/config"
	headscalePatterns "go.vervstack.ru/Velez/internal/patterns/headscale"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

const (
	Name = "headscale"

	groupName         = "verv_private_network"
	defaultConfigPath = "/etc/headscale/config.yaml"
)

type headscaleLauncher struct {
	ctx context.Context
	cfg config.Config

	clients node_clients.NodeClients
}

// LaunchHeadscale creates a headscale containe to connect to
func LaunchHeadscale(
	ctx context.Context,
	cfg config.Config,
	nodeClients node_clients.NodeClients,
) (*headscale.Client, error) {
	l := headscaleLauncher{ctx, cfg, nodeClients}

	isRunning, err := l.isServiceRunning()
	if err != nil {
		return nil, rerrors.Wrap(err, "error checking if headscale is running")
	}

	if !isRunning {
		err = l.deploy()
		if err != nil {
			return nil, rerrors.Wrap(err, "error deploying headscale")
		}
	}

	err = l.createContainerTask()
	if err != nil {
		return nil, rerrors.Wrap(err, "error starting headscale container")
	}

	client, err := headscale.New(ctx, nodeClients, Name)
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating headscale client")
	}

	return client, nil
}

func (l headscaleLauncher) isServiceRunning() (bool, error) {
	docker := l.clients.Docker()

	listReq := &velez_api.ListSmerds_Request{
		Limit: toolbox.ToPtr(uint32(1)),
		Name:  toolbox.ToPtr(headscalePatterns.ServiceName),
	}

	conts, err := docker.ListContainers(l.ctx, listReq)
	if err != nil {
		return false, rerrors.Wrap(err, "error listing containers")
	}

	if len(conts) == 0 {
		return false, nil
	}

	if conts[0].State == "running" {
		return true, nil
	}

	err = docker.Remove(l.ctx, conts[0].ID)
	if err != nil {
		return false, errors.Wrap(err, "error removing unhealthy container")
	}

	return false, nil
}

func (l headscaleLauncher) deploy() error {
	return nil
}

func (l headscaleLauncher) createContainerTask() error {
	//
	createContainerReq := headscalePatterns.Headscale(headscalePatterns.Settings{})

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
