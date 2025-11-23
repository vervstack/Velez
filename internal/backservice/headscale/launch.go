package headscale

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	errors "go.redsock.ru/rerrors"
	rtb "go.redsock.ru/toolbox"
	"go.redsock.ru/toolbox/keep_alive"

	"go.vervstack.ru/Velez/internal/backservice/env/container_service_task"
	"go.vervstack.ru/Velez/internal/clients"
	"go.vervstack.ru/Velez/internal/config"
)

const (
	Name         = "headscale"
	defaultImage = "docker.io/headscale/headscale:v0"
)

var initModeSync = sync.Once{}

func Launch(
	ctx context.Context,
	cfg config.Config,
	clients clients.NodeClients,
) {
	initModeSync.Do(func() {
		var err error
		err = launch(ctx, cfg, clients)
		if err != nil {
			logrus.Fatal(errors.Wrap(err))
		}
	})
}

func launch(
	ctx context.Context,
	cfg config.Config,
	nodeClients clients.NodeClients,
) error {
	taskConstructor := container_service_task.NewTaskRequest[struct{}]{
		ContainerName: Name,
		NodeClients:   nodeClients,

		ImageName: rtb.Coalesce(cfg.Environment.VpnServerImage, defaultImage),
		VolumeMounts: map[string][]string{
			"headscale": {
				"/etc/headscale",
				"/var/lib/headscale",
				"/var/run/headscale",
			},
		},
	}

	logrus.Info("Preparing HeadScale background task")

	headScaleTask, err := container_service_task.NewTask[struct{}](taskConstructor)
	if err != nil {
		return errors.Wrap(err, "error creating task")
	}
	// Launch
	keepAlive := keep_alive.KeepAlive(
		headScaleTask,
		keep_alive.WithCancel(ctx.Done()),
		keep_alive.WithCheckInterval(time.Second/2),
	)

	_ = keepAlive

	return nil
}
