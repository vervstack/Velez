package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/docker/docker/client"
	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"

	"github.com/godverv/Velez/internal/backservice/configuration"
	"github.com/godverv/Velez/internal/backservice/env"
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
	"github.com/godverv/Velez/pkg/velez_api"
	//_transport_imports
)

func main() {
	logrus.Println("starting app")

	// Core dependencies
	aCore := mustInitCore()

	// Verv Environment
	mustInitEnvironment(aCore)

	// Service layer
	serviceManager := mustInitServiceManager(aCore)

	// Back service
	initBackServices(aCore.cfg, serviceManager.GetContainerManagerService())

	// API
	mgr := mustInitAPI(aCore, serviceManager)

	err := mgr.Start(aCore.ctx)
	if err != nil {
		logrus.Fatalf("error starting api: %s", err)
	}

	ctx := context.Background()

	waitingForTheEnd()
	logrus.Println("shutting down the app")

	err = mgr.Stop(ctx)
	if err != nil {
		logrus.Errorf("error stopping api: %s", err)
	}

	if err = closer.Close(); err != nil {
		logrus.Fatalf("errors while shutting down application %s", err)
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

func mustInitServiceManager(aCore applicationCore) service.Services {
	matreshkaApi, err := grpcClients.NewMatreshkaBeAPIClient(aCore.ctx, aCore.cfg)
	if err != nil {
		logrus.Fatalf("error getting matreshka api: %s", err)
	}

	services, err := service_manager.New(aCore.ctx, aCore.cfg, aCore.dockerAPI, matreshkaApi)
	if err != nil {
		logrus.Fatalf("error creating service manager: %s", err)
	}

	if aCore.cfg.GetBool(config.ShutDownOnExit) {
		closer.Add(smerdsDropper(services.GetContainerManagerService()))
	}

	return services
}

func mustInitEnvironment(aCore applicationCore) {
	err := env.StartNetwork(aCore.dockerAPI)
	if err != nil {
		logrus.Fatalf("error creating network: %s", err)
	}

	conf := configuration.New(aCore.dockerAPI)
	err = conf.Start()
	if err != nil {
		logrus.Fatalf("error launching config backservice: %s", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	go cron.KeepAlive(ctx, conf)
	closer.Add(func() error { cancel(); return nil })
}

func initBackServices(cfg config.Config, cm service.ContainerManager) {
	ctx, c := context.WithCancel(context.Background())
	closer.Add(func() error {
		c()
		return nil
	})

	go cron.KeepAlive(ctx, watchtower.New(cfg, cm))

	if cfg.GetBool(config.PortainerEnabled) {
		go cron.KeepAlive(ctx, portainer.New(cm))
	}
}

func smerdsDropper(manager service.ContainerManager) func() error {
	return func() error {
		logrus.Infof("%s env variable is set to TRUE. Dropping launched smerds", config.ShutDownOnExit)
		logrus.Infof("Listing launched smerds")
		ctx := context.Background()

		smerds, err := manager.ListSmerds(ctx, &velez_api.ListSmerds_Request{})
		if err != nil {
			return err
		}

		b, err := yaml.Marshal(smerds.Smerds)
		if err != nil {
			b = []byte(fmt.Sprintf("%v", smerds.Smerds))
		}
		logrus.Infof("%d smerds is active: %v", len(smerds.Smerds), string(b))

		dropReq := &velez_api.DropSmerd_Request{
			Uuids: make([]string, len(smerds.Smerds)),
		}

		for i := range smerds.Smerds {
			dropReq.Uuids[i] = smerds.Smerds[i].Uuid
		}

		logrus.Infof("Dropping %d smerds", len(smerds.Smerds))

		dropSmerds, err := manager.DropSmerds(ctx, dropReq)
		if err != nil {
			return err
		}

		logrus.Infof("%d smerds dropped successfully", len(dropSmerds.Successful))
		if len(dropSmerds.Successful) != 0 {
			logrus.Infof("Dropped smerds: %v", dropSmerds.Successful)
		}

		if len(dropSmerds.Failed) != 0 {
			logrus.Errorf("%d smerds failed to drop", len(dropSmerds.Failed))
			for _, f := range dropSmerds.Failed {
				logrus.Errorf("error dropping %s. Cause %s", f.Uuid, f.Cause)
			}
		}

		return nil
	}
}

type applicationCore struct {
	ctx context.Context

	cfg config.Config

	dockerAPI       client.CommonAPIClient
	securityManager security.Manager
}

func mustInitCore() (c applicationCore) {
	var err error

	// Config
	{
		c.cfg, err = config.Load()
		if err != nil {
			logrus.Fatalf("error reading config %s", err.Error())
		}

	}

	// Startup ctx
	{
		if c.cfg.AppInfo().StartupDuration == 0 {
			logrus.Fatalf("no startup duration in config")
		}

		var cancel func()
		c.ctx, cancel = context.WithTimeout(context.Background(), c.cfg.AppInfo().StartupDuration)
		closer.Add(func() error { cancel(); return nil })
	}

	// Docker api
	{
		c.dockerAPI, err = docker.NewClient()
		if err != nil {
			logrus.Fatalf("erorr getting docker api client: %s", err)
		}
		closer.Add(c.dockerAPI.Close)
	}

	// Security access layer
	{
		c.securityManager = security.NewSecurityManager(c.cfg.GetString(config.CustomPassToKey))

		err = c.securityManager.Start()
		if err != nil {
			logrus.Fatalf("error starting security manager: %s", err)
		}

		closer.Add(c.securityManager.Stop)
	}

	return
}

func mustInitAPI(aCore applicationCore, services service.Services) transport.Server {
	mgr := transport.NewManager()

	grpcConf, err := aCore.cfg.Api().GRPC(config.ApiGrpc)
	if err != nil {
		logrus.Fatalf("error getting grpc from config: %s", err)
	}

	srv, err := grpc.NewServer(
		aCore.cfg,
		grpcConf,
		services,
		aCore.securityManager,
	)
	if err != nil {
		logrus.Fatalf("error creating grpc server: %s", err)
	}

	mgr.AddServer(srv)

	return mgr
}
