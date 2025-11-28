package headscale

import (
	"context"
	_ "embed"
	"sync"
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

	"go.vervstack.ru/Velez/internal/backservice/env/container_service_task"
	"go.vervstack.ru/Velez/internal/clients"
	"go.vervstack.ru/Velez/internal/config"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/pipelines"
)

const (
	Name         = "headscale"
	defaultImage = "headscale/headscale:v0.27.1"
)

//go:embed config.yaml
var defaultConfig []byte

var initModeSync = sync.Once{}

type launcher struct {
	ctx context.Context
	cfg config.Config

	clients clients.NodeClients
}

func Launch(
	ctx context.Context,
	cfg config.Config,
	clients clients.NodeClients,
) {
	initModeSync.Do(func() {
		l := launcher{ctx, cfg, clients}
		err := l.launch()
		if err != nil {
			logrus.Fatal(rerrors.Wrap(err))
		}
	})
}

func (l launcher) launch() (err error) {
	err = l.initVolume()
	if err != nil {
		return rerrors.Wrap(err, "error initiating volume for headscale")
	}

	err = l.copyConfigToVolume()
	if err != nil {
		return rerrors.Wrap(err, "error coping headscale config to volume")
	}

	err = l.startContainer()
	if err != nil {
		return rerrors.Wrap(err, "error starting headscale container")
	}

	return nil
}

func (l launcher) initVolume() error {
	apiClient := l.clients.Docker().Client()

	req := volume.CreateOptions{
		Labels: map[string]string{
			labels.CreatedWithVelezLabel: "true",
		},
		Name: Name,
	}

	_, err := apiClient.VolumeCreate(l.ctx, req)
	if err != nil {
		return rerrors.Wrap(err, "error creating volume for headscale")
	}

	return nil
}

func (l launcher) copyConfigToVolume() (err error) {
	copyToVolumeReq := domain.CopyToVolumeRequest{
		VolumeName: Name,
		PathToFiles: map[string][]byte{
			"/etc/headscale/config.yaml": defaultConfig,
		},
	}
	runner := pipelines.NewCopyToVolumeRunner(l.clients, copyToVolumeReq)
	err = runner.Run(l.ctx)
	if err != nil {
		return rerrors.Wrap(err, "error during to volume coping")
	}

	return nil
}

func (l launcher) startContainer() error {
	createContainerReq := container.CreateRequest{
		Config: &container.Config{
			Hostname: Name,
			ExposedPorts: nat.PortSet{
				"8080/tcp": struct{}{},
			},
			Cmd: strslice.StrSlice{"serve"},
			Healthcheck: &container.HealthConfig{
				Test: []string{"CMD", "headscale", "health"},
			},

			Image: rtb.Coalesce(l.cfg.Environment.VpnServerImage, defaultImage),

			Labels: map[string]string{
				labels.AutoUpgrade:           "true",
				labels.CreatedWithVelezLabel: "true",
			},
		},
		HostConfig: &container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeVolume,
					Source: Name,
					Target: "/etc/headscale",
				},
			},
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
