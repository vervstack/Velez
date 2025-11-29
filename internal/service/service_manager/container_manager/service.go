package container_manager

import (
	"github.com/docker/docker/client"

	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/service"
)

type ContainerManager struct {
	dockerWrapper node_clients.Docker
	dockerAPI     client.APIClient

	portManager node_clients.PortManager

	configService service.ConfigurationService
}

func New(
	internalClients node_clients.NodeClients,
	configurator service.ConfigurationService,
) *ContainerManager {

	return &ContainerManager{
		dockerAPI:   internalClients.Docker().Client(),
		portManager: internalClients.PortManager(),

		configService: configurator,
		dockerWrapper: internalClients.Docker(),
	}
}
