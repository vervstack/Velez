package container_manager_v1

import (
	"github.com/godverv/Velez/internal/clients"
	"github.com/godverv/Velez/internal/service"
)

const (
	CreatedWithVelezLabel = "CREATED_WITH_VELEZ"
)

type ContainerManager struct {
	docker clients.Docker

	portManager   clients.PortManager
	deployManager clients.DeployManager

	configService service.ConfigurationService
}

func NewContainerManager(
	internalClients clients.NodeClients,
	configurator service.ConfigurationService,
) *ContainerManager {

	return &ContainerManager{
		docker:        internalClients.Docker(),
		portManager:   internalClients.PortManager(),
		deployManager: internalClients.DeployManager(),

		configService: configurator,
	}
}
