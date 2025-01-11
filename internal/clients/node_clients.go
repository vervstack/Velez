package clients

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	errors "go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox/closer"

	"github.com/godverv/Velez/internal/clients/docker"
	"github.com/godverv/Velez/internal/clients/hardware"
	"github.com/godverv/Velez/internal/clients/ports"
	"github.com/godverv/Velez/internal/config"
	"github.com/godverv/Velez/internal/security"
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
		closer.Add(cls.docker.Close)

		var pong types.Ping
		pong, err = cls.docker.Ping(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "error pinging docker api client")
		}
		_ = pong
	}

	// Security access layer
	{
		if !cfg.Environment.DisableAPISecurity {
			logrus.Debug("Initializing security manager")

			cls.securityManager = security.NewSecurityManager(cfg.Environment.CustomPassToKey)

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

	return cls, nil
}

func (c *nodeClients) DockerAPI() client.CommonAPIClient {
	return c.docker
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
