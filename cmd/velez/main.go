package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/godverv/Velez/internal/backservice/portainer"
	"github.com/godverv/Velez/internal/backservice/security"
	"github.com/godverv/Velez/internal/backservice/watchtower"
	"github.com/godverv/Velez/internal/clients/docker"
	grpcClients "github.com/godverv/Velez/internal/clients/grpc"
	"github.com/godverv/Velez/internal/config"
	"github.com/godverv/Velez/internal/cron"
	"github.com/godverv/Velez/internal/service"
	"github.com/godverv/Velez/internal/service/service_manager"
	"github.com/godverv/Velez/internal/transport"
	"github.com/godverv/Velez/internal/transport/grpc"
	"github.com/godverv/Velez/internal/utils/closer"
	//_transport_imports
)

func main() {
	logrus.Println("starting app")

	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		logrus.Fatalf("error reading config %s", err.Error())
	}

	if cfg.AppInfo().StartupDuration == 0 {
		logrus.Fatalf("no startup duration in config")
	}

	ctx, cancel := context.WithTimeout(ctx, cfg.AppInfo().StartupDuration)
	defer cancel()

	// Security acces layer
	securityManager := security.NewSecurityManager("./velez/private.key")
	err = securityManager.Start()
	if err != nil {
		logrus.Fatalf("error starting security manager: %s", err)
	}
	closer.Add(securityManager.Stop)

	// Service layer
	serviceManager := mustInitContainerManagerService(ctx, cfg)

	// API
	mgr := transport.NewManager()
	{
		grpcConf, err := cfg.Api().GRPC(config.ApiGrpc)
		if err != nil {
			logrus.Fatalf("error getting grpc from config: %s", err)
		}

		srv, err := grpc.NewServer(
			cfg,
			grpcConf,
			serviceManager,
			security.GrpcInterceptor(securityManager),
		)
		if err != nil {
			logrus.Fatalf("error creating grpc server: %s", err)
		}

		mgr.AddServer(srv)
	}

	err = mgr.Start(ctx)
	if err != nil {
		logrus.Fatalf("error starting api: %s", err)
	}

	ctx = context.Background()

	initCron(ctx, cfg, serviceManager)

	waitingForTheEnd()
	logrus.Println("shutting down the app")

	err = mgr.Stop(ctx)
	if err != nil {
		logrus.Errorf("error stopping api: %s", err)
	}

	if err = closer.Close(); err != nil {
		logrus.Fatalf("errors while shutting down application %s", err.Error())
	}
}

// rscli comment: an obligatory function for tool to work properly.
// must be called in the main function above
// also this is an LP song name reference, so no rules can be applied to the function name
func waitingForTheEnd() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-done
}

func mustInitContainerManagerService(ctx context.Context, cfg config.Config) service.Services {
	dockerApi, err := docker.NewClient()
	if err != nil {
		logrus.Fatalf("erorr getting docker api client: %s", err)
	}

	matreshkaApi, err := grpcClients.NewMatreshkaBeAPIClient(ctx, cfg)
	if err != nil {
		logrus.Fatalf("error getting matreshka api: %s", err)
	}

	s, err := service_manager.New(cfg, dockerApi, matreshkaApi)
	if err != nil {
		logrus.Fatalf("error creating service manager: %s", err)
	}

	return s
}

func initCron(ctx context.Context, cfg config.Config, sm service.Services) {
	wt := watchtower.NewWatchTower(cfg, sm.GetContainerManagerService())
	go cron.KeepAlive(ctx, wt)
	closer.Add(wt.Kill)

	if cfg.GetBool(config.PortainerEnabled) {
		pt := portainer.NewPortainer(sm.GetContainerManagerService())
		go cron.KeepAlive(ctx, pt)
		closer.Add(pt.Kill)
	}
}
