package container_manager_v1

import (
	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/client"
	"github.com/godverv/matreshka-be/pkg/matreshka_api"

	"github.com/godverv/Velez/internal/config"
	"github.com/godverv/Velez/internal/service/service_manager/container_manager_v1/config_manager"
	"github.com/godverv/Velez/internal/service/service_manager/container_manager_v1/port_manager"
	"github.com/godverv/Velez/internal/service/service_manager/container_manager_v1/resource_manager"
)

type ContainerManager struct {
	docker client.CommonAPIClient

	portManager     *port_manager.PortManager
	configManager   *config_manager.Configurator
	resourceManager *resource_manager.ResourceManager

	isNodeModeOn bool
}

func NewContainerManager(
	cfg config.Config,

	docker client.CommonAPIClient,

	configClient matreshka_api.MatreshkaBeAPIClient,

	portManager *port_manager.PortManager,

) (c *ContainerManager, err error) {
	c = &ContainerManager{
		docker: docker,

		resourceManager: resource_manager.New(docker),

		isNodeModeOn: cfg.GetBool(config.NodeMode),

		portManager: portManager,
	}

	c.configManager, err = config_manager.New(configClient)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create config manager")
	}

	return c, nil
}
