package managers

import (
	"context"

	"github.com/Red-Sock/toolbox/closer"
	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"

	"github.com/godverv/Velez/internal/clients"
	"github.com/godverv/Velez/internal/clients/docker"
	"github.com/godverv/Velez/internal/clients/docker/deploy_manager"
	"github.com/godverv/Velez/internal/clients/hardware"
	"github.com/godverv/Velez/internal/clients/ports"
	"github.com/godverv/Velez/internal/clients/security"
	"github.com/godverv/Velez/internal/config"
)

type internalClients struct {
	docker *docker.Docker

	portManager     clients.PortManager
	hardwareManager clients.HardwareManager

	deployManager   clients.DeployManager
	securityManager clients.SecurityManager
}

func NewInternalClients(ctx context.Context, cfg config.Config) (clients.InternalClients, error) {
	var err error
	cls := &internalClients{}

	// Docker engine
	{
		logrus.Debug("Initializing docker client")
		cls.docker, err = docker.NewClient()
		if err != nil {
			return nil, errors.Wrap(err, "error getting docker api client")
		}
		closer.Add(cls.docker.Close)
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

func (c *internalClients) DockerAPI() client.CommonAPIClient {
	return c.docker
}

func (c *internalClients) Docker() clients.Docker {
	return c.docker
}

func (c *internalClients) DeployManager() clients.DeployManager {
	return c.deployManager
}

func (c *internalClients) PortManager() clients.PortManager {
	return c.portManager
}

func (c *internalClients) HardwareManager() clients.HardwareManager {
	return c.hardwareManager
}

func (c *internalClients) SecurityManager() clients.SecurityManager {
	return c.securityManager
}
