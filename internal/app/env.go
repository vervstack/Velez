package app

import (
	"time"

	"github.com/Red-Sock/toolbox/keep_alive"
	errors "github.com/Red-Sock/trace-errors"
	"github.com/godverv/makosh/pkg/makosh_be"
	"github.com/sirupsen/logrus"
	grpc2 "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/godverv/Velez/internal/backservice/configuration"
	"github.com/godverv/Velez/internal/backservice/env"
	"github.com/godverv/Velez/internal/backservice/service_discovery_task"
	"github.com/godverv/Velez/internal/clients/grpc"
	"github.com/godverv/Velez/internal/clients/makosh"
)

func (c *Custom) initEnvironment(a *App) {
	err := env.StartNetwork(c.InternalClients.Docker())
	if err != nil {
		logrus.Fatalf("error creating network: %s", err)
	}

	err = env.StartVolumes(c.InternalClients.Docker())
	if err != nil {
		logrus.Fatalf("error creating volumes %s", err)
	}

	if !a.Cfg.Environment.NodeMode {
		return
	}

	c.setupServiceDiscovery(a)

	matreshkaTask, err := configuration.New(a.Cfg, c.InternalClients)
	if err != nil {
		logrus.Fatalf("error creating configuration background task %s", errors.Wrap(err))
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
	_, err = a.GrpcMakosh.UpsertEndpoints(a.Ctx, metreshkaEndpoint)
	if err != nil {
		logrus.Fatalf("error upserting endpoint for matreshka %s", err)
	}

}

func (c *Custom) setupServiceDiscovery(a *App) {
	makoshBackgroundTask, err := service_discovery_task.New(a.Cfg, c.InternalClients)
	if err != nil {
		logrus.Fatalf("error creating service discovery background task: %s", err)
	}

	logrus.Info("Starting service discovery background task")
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

	a.GrpcMakosh, err = makosh.New(a.Cfg, makoshBackgroundTask.AuthToken,
		grpc2.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logrus.Fatalf("error creating makosh client %s", errors.Wrap(err))
	}

	err = grpc.RegisterServiceDiscovery(a.Cfg)
	if err != nil {
		logrus.Fatalf("error initializing service discovery %s", err)
	}
}
