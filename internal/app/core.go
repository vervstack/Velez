package app

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/godverv/Velez/internal/backservice/security"
	"github.com/godverv/Velez/internal/clients/docker"
	"github.com/godverv/Velez/internal/config"
	"github.com/godverv/Velez/internal/service/service_manager/container_manager_v1/port_manager"
	"github.com/godverv/Velez/internal/utils/closer"
)

func (a *App) mustInitCore() {
	var err error

	if a.Ctx == nil {
		a.Ctx = context.Background()
	}

	var cancel context.CancelFunc
	a.Ctx, cancel = context.WithCancel(a.Ctx)
	closer.Add(func() error {
		cancel()
		return nil
	})

	// Load config
	{
		a.Cfg, err = config.Load()
		if err != nil {
			logrus.Fatalf("error reading config %s", err.Error())
		}

	}

	// Docker api
	{
		a.Docker, err = docker.NewClient()
		if err != nil {
			logrus.Fatalf("erorr getting docker api client: %s", err)
		}
		closer.Add(a.Docker.Close)
	}

	// Security access layer
	if !a.Cfg.GetEnvironment().DisableAPISecurity {
		a.SecurityManager = security.NewSecurityManager(a.Cfg.GetEnvironment().CustomPassToKey)

		err = a.SecurityManager.Start()
		if err != nil {
			logrus.Fatalf("error starting security manager: %s", err)
		}

		closer.Add(a.SecurityManager.Stop)
	}

	// port manager
	{
		a.PortManager, err = port_manager.NewPortManager(context.Background(), a.Cfg, a.Docker)
		if err != nil {
			logrus.Fatalf("error creating port manager %s", err)
		}
	}
	return
}
