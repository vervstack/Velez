package clients

import (
	"context"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/client"
	"github.com/godverv/matreshka-be/pkg/matreshka_api"
	"github.com/sirupsen/logrus"

	"github.com/godverv/Velez/internal/clients/configurator"
	"github.com/godverv/Velez/internal/clients/docker"
	"github.com/godverv/Velez/internal/clients/docker/deploy_manager"
	grpcClients "github.com/godverv/Velez/internal/clients/grpc"
	"github.com/godverv/Velez/internal/clients/hardware"
	"github.com/godverv/Velez/internal/clients/ports"
	"github.com/godverv/Velez/internal/clients/security"
	"github.com/godverv/Velez/internal/config"
	"github.com/godverv/Velez/internal/utils/closer"
)

type clients struct {
	docker    *docker.Docker
	matreshka matreshka_api.MatreshkaBeAPIClient

	portManager     PortManager
	hardwareManager HardwareManager
	securityManager security.Manager

	configurator  Configurator
	deployManager DeployManager
}

func New(ctx context.Context, cfg config.Config) (Clients, error) {
	var err error
	cls := &clients{}

	// Docker engine
	{
		cls.docker, err = docker.NewClient()
		if err != nil {
			return nil, errors.Wrap(err, "error getting docker api client")
		}
		closer.Add(cls.docker.Close)
	}
	// Matreshka
	{
		cls.matreshka, err = grpcClients.NewMatreshkaBeAPIClient(ctx, cfg)
		if err != nil {
			logrus.Fatalf("error getting matreshka api: %s", err)
		}
	}

	// Security access layer
	{
		if !cfg.GetEnvironment().DisableAPISecurity {
			cls.securityManager = security.NewSecurityManager(cfg.GetEnvironment().CustomPassToKey)

			err = cls.securityManager.Start()
			if err != nil {
				logrus.Fatalf("error starting security manager: %s", err)
			}

			closer.Add(cls.securityManager.Stop)
		}
	}

	// Port manager
	{
		cls.portManager, err = ports.NewPortManager(ctx, cfg, cls.docker)
		if err != nil {
			logrus.Fatalf("error creating port manager %s", err)
		}
	}

	// Configurator
	{
		cls.configurator = configurator.New(cls.matreshka, cls.docker)
	}

	// Hardware
	{
		cls.hardwareManager = hardware.New()
	}

	// Deploy
	{
		cls.deployManager = deploy_manager.New(cls.docker)
	}
	return cls, nil
}

func (c *clients) DockerAPI() client.CommonAPIClient {
	return c.docker
}

func (c *clients) Docker() Docker {
	return c.docker
}

func (c *clients) Configurator() Configurator {
	return c.configurator
}

func (c *clients) DeployManager() DeployManager {
	return c.deployManager
}

func (c *clients) PortManager() PortManager {
	return c.portManager
}

func (c *clients) HardwareManager() HardwareManager {
	return c.hardwareManager
}

func (c *clients) SecurityManager() security.Manager {
	return c.securityManager
}
