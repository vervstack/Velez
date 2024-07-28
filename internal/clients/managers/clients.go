package managers

import (
	"context"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/client"
	"github.com/godverv/matreshka-be/pkg/matreshka_api"
	"github.com/sirupsen/logrus"

	"github.com/godverv/Velez/internal/backservice/service_discovery"
	"github.com/godverv/Velez/internal/clients"
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

type clientsManager struct {
	docker    *docker.Docker
	matreshka matreshka_api.MatreshkaBeAPIClient

	portManager     clients.PortManager
	hardwareManager clients.HardwareManager
	securityManager clients.SecurityManager

	configurator     clients.Configurator
	deployManager    clients.DeployManager
	serviceDiscovery clients.ServiceDiscovery
}

func New(ctx context.Context, cfg config.Config, sd clients.ServiceDiscovery) (clients.Clients, error) {
	var err error
	cls := &clientsManager{
		serviceDiscovery: sd,
	}

	// Docker engine
	{
		logrus.Debug("Initializing docker client")
		cls.docker, err = docker.NewClient()
		if err != nil {
			return nil, errors.Wrap(err, "error getting docker api client")
		}
		closer.Add(cls.docker.Close)
	}
	// Matreshka
	{
		logrus.Debug("Initializing matreshka client")
		cls.matreshka, err = grpcClients.NewMatreshkaBeAPIClient(ctx, cfg)
		if err != nil {
			logrus.Fatalf("error getting matreshka api: %s", err)
		}
	}

	{
		logrus.Debug("Initializing makosh client")
		cls.serviceDiscovery, err = service_discovery.New()
	}

	// Security access layer
	{
		if !cfg.GetEnvironment().DisableAPISecurity {
			logrus.Debug("Initializing security manager")

			cls.securityManager = security.NewSecurityManager(cfg.GetEnvironment().CustomPassToKey)

			err = cls.securityManager.Start()
			if err != nil {
				logrus.Fatalf("error starting security manager: %s", err)
			}

			closer.Add(cls.securityManager.Stop)
		} else {
			logrus.Debug("Security manager disabled")
		}
	}

	// Port manager
	{
		logrus.Debug("Initializing port manager")

		cls.portManager, err = ports.NewPortManager(ctx, cfg, cls.docker)
		if err != nil {
			logrus.Fatalf("error creating port manager %s", err)
		}
	}

	// Configurator
	{
		logrus.Debug("Initializing configuration manager")
		cls.configurator = configurator.New(cls.matreshka, cls.docker)
	}

	// Hardware
	{
		logrus.Debug("Initializing hardware manager")
		cls.hardwareManager = hardware.New()
	}

	// Deploy
	{
		logrus.Debug("Initializing deployment manager")
		cls.deployManager = deploy_manager.New(cls.docker)
	}

	return cls, nil
}

func (c *clientsManager) DockerAPI() client.CommonAPIClient {
	return c.docker
}

func (c *clientsManager) Docker() clients.Docker {
	return c.docker
}

func (c *clientsManager) Configurator() clients.Configurator {
	return c.configurator
}

func (c *clientsManager) DeployManager() clients.DeployManager {
	return c.deployManager
}

func (c *clientsManager) PortManager() clients.PortManager {
	return c.portManager
}

func (c *clientsManager) HardwareManager() clients.HardwareManager {
	return c.hardwareManager
}

func (c *clientsManager) SecurityManager() clients.SecurityManager {
	return c.securityManager
}

func (c *clientsManager) ServiceDiscovery() clients.ServiceDiscovery {
	return c.serviceDiscovery
}
