package node_clients

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/sirupsen/logrus"
	errors "go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox/closer"

	"go.vervstack.ru/Velez/internal/clients/node_clients/docker"
	"go.vervstack.ru/Velez/internal/clients/node_clients/hardware"
	"go.vervstack.ru/Velez/internal/clients/node_clients/local_state"
	"go.vervstack.ru/Velez/internal/clients/node_clients/ports"
	"go.vervstack.ru/Velez/internal/config"
)

// NodeClients - container for node level clients
type NodeClients interface {
	// Docker - returns basic DockerEngine API
	Docker() Docker

	PortManager() PortManager
	PortManagerContainer() *ports.Container
	LocalStateManager() StateManager

	HardwareManager() HardwareManager
}

type nodeClients struct {
	docker *docker.Docker

	portManager     *ports.Container
	hardwareManager HardwareManager

	securityManager StateManager
}

func NewNodeClients(ctx context.Context, cfg config.Config) (NodeClients, error) {
	var err error
	cls := &nodeClients{}

	// Docker engine
	{
		logrus.Debug("Initializing docker client")
		cls.docker, err = docker.NewClient(cfg.Environment.CustomLabels, cfg.Environment.ContainerSuffix)
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

			cls.securityManager = local_state.NewSecurityManager(cfg)

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

		usedPorts, listErr := cls.docker.ListOccupiedPorts(ctx)
		if listErr != nil {
			logrus.Error("error listing occupied ports")
		}

		portManager := ports.NewPortManager(cfg.Environment.AvailablePorts, usedPorts)
		cls.portManager = ports.NewContainer(portManager)
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

func (c *nodeClients) PortManagerContainer() *ports.Container {
	return c.portManager
}

func (c *nodeClients) HardwareManager() HardwareManager {
	return c.hardwareManager
}

func (c *nodeClients) LocalStateManager() StateManager {
	return c.securityManager
}
