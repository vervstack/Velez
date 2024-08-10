package app

import (
	"time"

	"github.com/Red-Sock/toolbox/keep_alive"
	"github.com/godverv/makosh/pkg/makosh_be"
	"github.com/sirupsen/logrus"

	"github.com/godverv/Velez/internal/backservice/configuration"
	"github.com/godverv/Velez/internal/backservice/env"
	"github.com/godverv/Velez/internal/backservice/service_discovery_task"
	"github.com/godverv/Velez/internal/clients/grpc"
	"github.com/godverv/Velez/internal/clients/makosh"
)

func (a *App) MustInitEnvironment() {
	err := env.StartNetwork(a.InternalClients.Docker())
	if err != nil {
		logrus.Fatalf("error creating network: %s", err)
	}

	err = env.StartVolumes(a.InternalClients.Docker())
	if err != nil {
		logrus.Fatalf("error creating volumes %s", err)
	}

	envVars := a.Cfg.GetEnvironment()

	if envVars.NodeMode {
		a.setupServiceDiscovery()

		matreshkaTask, err := configuration.New(a.Cfg, a.InternalClients)
		if err != nil {
			logrus.Fatalf("error creating configuration background task %s", err)
		}
		go keep_alive.KeepAlive(matreshkaTask, keep_alive.WithCancel(a.Ctx.Done()))

		metreshkaEndpoint := &makosh_be.UpsertEndpoints_Request{
			Endpoints: []*makosh_be.Endpoint{
				{
					ServiceName: "matreshka",
					Addrs:       []string{},
				},
			},
		}
		_, err = a.MakoshClient.UpsertEndpoints(a.Ctx, metreshkaEndpoint)
		if err != nil {
			logrus.Fatalf("error upserting endpoint for matreshka %s", err)
		}
	}

}

func (a *App) setupServiceDiscovery() {
	makoshBackgroundTask, err := service_discovery_task.New(a.Cfg, a.InternalClients)
	if err != nil {
		logrus.Fatalf("error creating service discovery background task: %s", err)
	}
	go keep_alive.KeepAlive(makoshBackgroundTask, keep_alive.WithCancel(a.Ctx.Done()))

	t := time.NewTicker(time.Second * 2)
	for range t.C {
		isAlive, err := makoshBackgroundTask.IsAlive()
		if err != nil {
			logrus.Fatalf("error during setting up makosh service %s", err)
		}

		if isAlive {
			t.Stop()
			break
		}
	}

	a.MakoshClient, err = makosh.New(a.Cfg, makoshBackgroundTask.AuthToken)
	if err != nil {
		logrus.Fatalf("error creating makosh client %s", err)
	}

	err = grpc.RegisterServiceDiscovery(a.Cfg)
	if err != nil {
		logrus.Fatalf("error initializing service discovery %s", err)
	}
}
