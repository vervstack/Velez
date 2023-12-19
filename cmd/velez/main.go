package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/godverv/Velez/internal/client/docker"
	"github.com/godverv/Velez/internal/config"
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
	closer.Add(func() error {
		cancel()
		return nil
	})

	mgr := transport.NewManager()

	{
		dockerApi, err := docker.NewClient()
		if err != nil {
			logrus.Fatalf("erorr getting docker api client: %s", err)
		}

		grpcConf, err := cfg.Api().GRPC(config.ApiGrpc)
		if err != nil {
			logrus.Fatalf("error getting grpc from config: %s", err)
		}

		srv, err := grpc.NewServer(cfg, grpcConf, dockerApi)
		if err != nil {
			logrus.Fatalf("error creating grpc server: %s", err)
		}

		mgr.AddServer(srv)
	}

	err = mgr.Start(ctx)
	if err != nil {
		logrus.Fatalf("error starting api: %s", err)
	}

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
