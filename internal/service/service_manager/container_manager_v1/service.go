package container_manager_v1

import (
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
	volumesPath  string
}

func NewContainerManager(
	cfg config.Config,

	docker client.CommonAPIClient,

	configClient matreshka_api.MatreshkaBeAPIClient,

	portManager *port_manager.PortManager,

) (c *ContainerManager, err error) {
	c = &ContainerManager{
		docker:          docker,
		configManager:   config_manager.New(configClient, cfg),
		resourceManager: resource_manager.New(docker),

		isNodeModeOn: cfg.GetBool(config.NodeMode),
		volumesPath:  cfg.GetString(config.SmerdVolumePath),

		portManager: portManager,
	}

	return c, nil
}
