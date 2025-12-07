package verv_private_network

import (
	"context"
	_ "embed"
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
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/pipelines"
)

const (
	Name = "headscale"

	groupName         = "verv_private_network"
	defaultImage      = "headscale/headscale:0.27.2-rc.1"
	defaultPort       = nat.Port("8080/tcp")
	defaultConfigPath = "/etc/headscale/config.yaml"
)

var (
	//go:embed config.yaml
	defaultConfig []byte
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

	err := l.initVolume()
	if err != nil {
		return nil, rerrors.Wrap(err, "error initiating volume for headscale")
	}

	err = l.copyConfigToVolume()
	if err != nil {
		return nil, rerrors.Wrap(err, "error coping headscale config to volume")
	}

	err = l.startContainer()
	if err != nil {
		return nil, rerrors.Wrap(err, "error starting headscale container")
	}

	client, err := headscale.New(ctx, nodeClients, Name)
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating headscale client")
	}

	return client, nil
}

func (l headscaleLauncher) initVolume() error {
	apiClient := l.clients.Docker().Client()

	req := volume.CreateOptions{
		// TODO add cluster name
		Name: Name + "-",
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

func (l headscaleLauncher) copyConfigToVolume() (err error) {
	copyToVolumeReq := domain.CopyToVolumeRequest{
		VolumeName: Name,
		PathToFiles: map[string][]byte{
			defaultConfigPath: defaultConfig,
		},
	}

	runner := pipelines.NewCopyToVolumeRunner(l.clients, copyToVolumeReq)

	err = runner.Run(l.ctx)
	if err != nil {
		return rerrors.Wrap(err, "error during to volume coping")
	}

	return nil
}

func (l headscaleLauncher) startContainer() error {
	createContainerReq := container.CreateRequest{
		Config: &container.Config{
			Hostname: Name,
			ExposedPorts: nat.PortSet{
				defaultPort: struct{}{},
			},
			Cmd: strslice.StrSlice{"serve"},
			Healthcheck: &container.HealthConfig{
				Test: []string{"CMD", "headscale", "health"},
			},

			Image: rtb.Coalesce(l.cfg.Environment.VpnServerImage, defaultImage),

			Labels: map[string]string{
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
