package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/godverv/Velez/internal/backservice"
	"github.com/godverv/Velez/internal/client/docker"
	"github.com/godverv/Velez/internal/config"
	"github.com/godverv/Velez/internal/cron"
	"github.com/godverv/Velez/internal/service"
	"github.com/godverv/Velez/internal/service/container_manager_v1"
	"github.com/godverv/Velez/internal/service/hardware_manager_v1"
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

	ctx, _ = context.WithTimeout(ctx, cfg.AppInfo().StartupDuration)

	cm, hm := mustInitContainerManagerService()

	mgr := transport.NewManager()
	{
		grpcConf, err := cfg.Api().GRPC(config.ApiGrpc)
		if err != nil {
			logrus.Fatalf("error getting grpc from config: %s", err)
		}

		srv, err := grpc.NewServer(cfg, grpcConf, cm, hm)
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
	ctx, cancel := context.WithCancel(ctx)
	closer.Add(func() error {
		cancel()
		return nil
	})

	initCron(ctx, cfg, cm)

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

func mustInitContainerManagerService() (service.ContainerManager, service.HardwareManager) {
	dockerApi, err := docker.NewClient()
	if err != nil {
		logrus.Fatalf("erorr getting docker api client: %s", err)
	}

	cm, err := container_manager_v1.NewContainerManager(dockerApi)
	if err != nil {
		logrus.Fatalf("error creating container manager")
	}

	return cm, hardware_manager_v1.New()
}

func initCron(ctx context.Context, cfg config.Config, cm service.ContainerManager) {
	go cron.KeepAlive(ctx, backservice.NewWatchTower(cfg, cm))
}
