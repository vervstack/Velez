package clients

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/sirupsen/logrus"
	errors "go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox/closer"

	"go.vervstack.ru/Velez/internal/clients/docker"
	"go.vervstack.ru/Velez/internal/clients/hardware"
	"go.vervstack.ru/Velez/internal/clients/ports"
	"go.vervstack.ru/Velez/internal/config"
	"go.vervstack.ru/Velez/internal/middleware/security"
)

// NodeClients - container for node level clients
type NodeClients interface {
	// Docker - returns basic DockerEngine API
	Docker() Docker

	PortManager() PortManager
	SecurityManager() SecurityManager

	HardwareManager() HardwareManager
}

type nodeClients struct {
	docker *docker.Docker

	portManager     PortManager
	hardwareManager HardwareManager

	securityManager SecurityManager
}

func NewNodeClientsContainer(ctx context.Context, cfg config.Config) (NodeClients, error) {
	var err error
	cls := &nodeClients{}

	// Docker engine
	{
		logrus.Debug("Initializing docker client")
		cls.docker, err = docker.NewClient()
		if err != nil {
			return nil, errors.Wrap(err, "error getting docker api client")
		}

		var pong types.Ping
		pong, err = cls.docker.Client().Ping(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "Can't ping docker api. If you are running Velez inside a container please provide docker socket via volume flag: -v /var/run/docker.sock:/var/run/docker.sock")
		}
		_ = pong
	}

	// Security access layer
	{
		if !cfg.Environment.DisableAPISecurity {
			logrus.Debug("Initializing security manager")

			cls.securityManager = security.NewSecurityManager(cfg)

			err = cls.securityManager.Start()
			if err != nil {
				logrus.Fatalf("error starting security manager: %s", err)
			}

			closer.Add(cls.securityManager.Stop)
		} else {
			logrus.Fatalf("Security manager disabled")
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

	return cls, nil
}

func (c *nodeClients) Docker() Docker {
	return c.docker
}

func (c *nodeClients) PortManager() PortManager {
	return c.portManager
}

func (c *nodeClients) HardwareManager() HardwareManager {
	return c.hardwareManager
}

func (c *nodeClients) SecurityManager() SecurityManager {
	return c.securityManager
}
