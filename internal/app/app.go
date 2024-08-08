package app

import (
	"context"

	rtb "github.com/Red-Sock/toolbox"
	"github.com/Red-Sock/toolbox/closer"
	errors "github.com/Red-Sock/trace-errors"
	"github.com/godverv/makosh/pkg/makosh_be"
	"github.com/sirupsen/logrus"

	"github.com/godverv/Velez/internal/clients"
	"github.com/godverv/Velez/internal/config"
	"github.com/godverv/Velez/internal/service"
	"github.com/godverv/Velez/internal/transport"
	"github.com/godverv/Velez/internal/transport/grpc"
)

type App struct {
	// Essentials
	Ctx                 context.Context
	Cfg                 config.Config
	ServiceDiscoveryURL string

	// Host communication and external services
	InternalClients clients.InternalClients
	ExternalClients clients.ExternalClients

	MakoshClient makosh_be.MakoshBeAPIClient

	// Business logic
	Services service.Services

	// Api
	GrpcApi       *grpc.Server
	ServerManager *transport.ServersManager
}

func New() (a *App) {
	logrus.Println("starting app")
	a = &App{}

	// Core dependencies
	a.MustInitCore()

	// Verv Environment
	a.MustInitEnvironment()

	// External clients
	a.MustInitExternalClients()

	// Service layer
	a.InitServiceManager()

	// API
	a.MustInitAPI()

	return a
}

func (a *App) Start() error {
	ctx := context.Background()
	err := a.ServerManager.Start(ctx)
	if err != nil {
		return errors.Wrap(err, "error starting api")
	}

	rtb.WaitForInterrupt()

	logrus.Println("shutting down the app")

	err = a.ServerManager.Stop(ctx)
	if err != nil {
		return errors.Wrap(err, "error stopping api")
	}

	err = closer.Close()
	if err != nil {
		return errors.Wrap(err, "errors while shutting down application")
	}

	return nil
}
