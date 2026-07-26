package container_manager

import (
	"github.com/docker/docker/client"

	"go.vervstack.ru/Velez/internal/clients/node_clients"
)

type ContainerManager struct {
	dockerWrapper node_clients.Docker
	dockerAPI     client.APIClient
}

func New(
	internalClients node_clients.NodeClients,
) *ContainerManager {
	return &ContainerManager{
		dockerAPI: internalClients.Docker().Client(),

		dockerWrapper: internalClients.Docker(),
	}
}
