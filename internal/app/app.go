package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/client"
	"github.com/godverv/matreshka-be/pkg/matreshka_api"
	"github.com/sirupsen/logrus"

	"github.com/godverv/Velez/internal/backservice/security"
	"github.com/godverv/Velez/internal/config"
	"github.com/godverv/Velez/internal/service"
	"github.com/godverv/Velez/internal/service/service_manager/container_manager_v1/port_manager"
	"github.com/godverv/Velez/internal/transport"
	"github.com/godverv/Velez/internal/transport/grpc"
	"github.com/godverv/Velez/internal/utils/closer"
)

type App struct {
	// Essentials
	Ctx context.Context
	Cfg config.Config

	// Host communication
	Docker          client.CommonAPIClient
	SecurityManager security.Manager
	PortManager     *port_manager.PortManager

	// Business logic
	Services service.Services

	// External services
	MatreshkaClient matreshka_api.MatreshkaBeAPIClient

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

	// Service layer
	a.MustInitServiceManager()

	// Back service
	a.InitBackServices()

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

	waitingForTheEnd()
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

// rscli comment: an obligatory function for tool to work properly.
// must be called in the main function above
// also this is an LP song name reference, so no rules can be applied to the function name
func waitingForTheEnd() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-done
}
