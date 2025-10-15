package container_manager

import (
	"github.com/docker/docker/client"

	"go.vervstack.ru/Velez/internal/clients"
	"go.vervstack.ru/Velez/internal/service"
)

type ContainerManager struct {
	dockerWrapper clients.Docker
	dockerAPI     client.APIClient

	portManager clients.PortManager

	configService service.ConfigurationService
}

func NewContainerManager(
	internalClients clients.NodeClients,
	configurator service.ConfigurationService,
) *ContainerManager {

	return &ContainerManager{
		dockerAPI:   internalClients.Docker().Client(),
		portManager: internalClients.PortManager(),

		configService: configurator,
		dockerWrapper: internalClients.Docker(),
	}
}
