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
	clients clients.Clients,
) *ContainerManager {

	return &ContainerManager{
		docker: clients.Docker(),

		portManager:   clients.PortManager(),
		configManager: clients.Configurator(),
		deployManager: clients.DeployManager(),
	}
}
