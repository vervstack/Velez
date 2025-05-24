package container_manager

import (
	"go.vervstack.ru/Velez/internal/clients"
	"go.vervstack.ru/Velez/internal/service"
)

type ContainerManager struct {
	docker clients.Docker

	portManager clients.PortManager

	configService service.ConfigurationService
}

func NewContainerManager(
	internalClients clients.NodeClients,
	configurator service.ConfigurationService,
) *ContainerManager {

	return &ContainerManager{
		docker:      internalClients.Docker(),
		portManager: internalClients.PortManager(),

		configService: configurator,
	}
}
