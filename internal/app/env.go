package app

import (
	"github.com/Red-Sock/toolbox/keep_alive"
	"github.com/sirupsen/logrus"

	"github.com/godverv/Velez/internal/backservice/configuration"
	"github.com/godverv/Velez/internal/backservice/env"
	"github.com/godverv/Velez/internal/backservice/service_discovery"
)

func (a *App) MustInitEnvironment() {
	err := env.StartNetwork(a.Clients.Docker())
	if err != nil {
		logrus.Fatalf("error creating network: %s", err)
	}

	err = env.StartVolumes(a.Clients.Docker())
	if err != nil {
		logrus.Fatalf("error creating volumes %s", err)
	}

	envVars := a.Cfg.GetEnvironment()

	if !envVars.NodeMode {
		return
	}

	matreshkaTask, err := configuration.New(a.Cfg, a.Clients)
	if err != nil {
		logrus.Fatalf("error creating configuration background task %s", err)
	}

	go keep_alive.KeepAlive(matreshkaTask, keep_alive.WithCancel(a.Ctx.Done()))

	makoshBackgroundTask, err := service_discovery.New(string(a.MakoshKey), a.Cfg, a.Clients)
	if err != nil {
		logrus.Fatalf("error creating service discovery background task: %s", err)
	}

	go keep_alive.KeepAlive(makoshBackgroundTask, keep_alive.WithCancel(a.Ctx.Done()))
}
