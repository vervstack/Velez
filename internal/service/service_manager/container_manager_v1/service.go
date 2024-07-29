package container_manager_v1

import (
	"github.com/godverv/Velez/internal/clients"
)

const (
	CreatedWithVelezLabel = "CREATED_WITH_VELEZ"
)

type ContainerManager struct {
	docker clients.Docker

	configManager clients.Configurator
	portManager   clients.PortManager
	deployManager clients.DeployManager
}

func NewContainerManager(
	internalClients clients.InternalClients,
	externalClients clients.ExternalClients,
) *ContainerManager {

	return &ContainerManager{
		docker: internalClients.Docker(),

		portManager:   internalClients.PortManager(),
		deployManager: internalClients.DeployManager(),

		configManager: externalClients.Configurator(),
	}
}
