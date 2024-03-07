package container_manager_v1

import (
	"context"

	"github.com/docker/docker/client"
	"github.com/godverv/matreshka-be/pkg/matreshka_api"
	"github.com/pkg/errors"

	"github.com/godverv/Velez/internal/config"
	"github.com/godverv/Velez/internal/service/container_manager_v1/config_manager"
	"github.com/godverv/Velez/internal/service/container_manager_v1/port_manager"
)

type ContainerManager struct {
	docker        client.CommonAPIClient
	portManager   *port_manager.PortManager
	configManager *config_manager.Configurator
}

func NewContainerManager(
	ctx context.Context,
	cfg config.Config,
	docker client.CommonAPIClient,
	configClient matreshka_api.MatreshkaBeAPIClient,
) (c *ContainerManager, err error) {
	c = &ContainerManager{
		docker:        docker,
		configManager: config_manager.New(configClient),
	}

	c.portManager, err = port_manager.NewPortManager(ctx, cfg, docker)
	if err != nil {
		return nil, errors.Wrap(err, "error creating port manager")
	}

	return c, nil
}
