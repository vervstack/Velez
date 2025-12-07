package verv_private_network

import (
	"context"
	"path"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/go-connections/nat"
	"github.com/sirupsen/logrus"
	"go.redsock.ru/rerrors"
	rtb "go.redsock.ru/toolbox"
	"go.redsock.ru/toolbox/keep_alive"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients/headscale"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/cluster/env/container_service_task"
	"go.vervstack.ru/Velez/internal/config"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

type tailscaleLauncher struct {
	ctx context.Context
	cfg config.Config

	clients node_clients.NodeClients
}

// LaunchTailscale creates a tailscale container - client for vpn
func LaunchTailscale(
	ctx context.Context,
	cfg config.Config,
	clients node_clients.NodeClients,
) (*headscale.Client, error) {

	l := tailscaleLauncher{ctx, cfg, clients}

	err := l.initVolume()
	if err != nil {
		return nil, rerrors.Wrap(err, "error initiating volume for headscale")
	}

	err = l.startContainer()
	if err != nil {
		return nil, rerrors.Wrap(err, "error starting headscale container")
	}

	client, err := headscale.New(ctx, clients, Name)
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating headscale client")
	}

	return client, nil
}

func (l tailscaleLauncher) initVolume() error {
	apiClient := l.clients.Docker().Client()

	req := volume.CreateOptions{
		Name: Name,
		Labels: map[string]string{
			labels.VervServiceLabel:  "true",
			labels.ComposeGroupLabel: groupName,
		},
	}

	_, err := apiClient.VolumeCreate(l.ctx, req)
	if err != nil {
		return rerrors.Wrap(err, "error creating volume for headscale")
	}

	return nil
}

func (l tailscaleLauncher) startContainer() error {
	createContainerReq := container.CreateRequest{
		Config: &container.Config{
			// TODO add cluster name
			Hostname: Name + "-",
			ExposedPorts: nat.PortSet{
				defaultPort: struct{}{},
			},
			Cmd: strslice.StrSlice{"serve"},
			Healthcheck: &container.HealthConfig{
				Test: []string{"CMD", "headscale", "health"},
			},

			Image: rtb.Coalesce(l.cfg.Environment.VpnServerImage, defaultImage),

			Labels: map[string]string{
				labels.AutoUpgrade:      "true",
				labels.VervServiceLabel: "true",
			},
		},
		HostConfig: &container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeVolume,
					Source: Name,
					Target: path.Dir(defaultConfigPath),
				},
				{
					Type:   mount.TypeVolume,
					Source: Name,
					Target: "/var/lib/headscale",
				},
			},
			PortBindings: map[nat.Port][]nat.PortBinding{},
		},
	}

	// TODO implement exposure via config
	createContainerReq.HostConfig.PortBindings[defaultPort] = []nat.PortBinding{
		{
			HostPort: "8080",
		},
	}

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
