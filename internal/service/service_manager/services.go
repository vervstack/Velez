package service_manager

import (
	"github.com/godverv/Velez/internal/clients"
	"github.com/godverv/Velez/internal/service"
	"github.com/godverv/Velez/internal/service/service_manager/configurator"
	"github.com/godverv/Velez/internal/service/service_manager/container_manager_v1"
	"github.com/godverv/Velez/internal/service/service_manager/smerd_launcher"
)

type ServiceManager struct {
	*container_manager_v1.ContainerManager

	*smerd_launcher.SmerdLauncher

	*configurator.Configurator
}

func New(
	nodeClients clients.NodeClients,
	clusterClients clients.ClusterClients,
) service.Services {

	configService := configurator.New(
		clusterClients.Configurator(),
		nodeClients.Docker(),
	)

	return &ServiceManager{
		Configurator: configService,

		ContainerManager: container_manager_v1.NewContainerManager(nodeClients, configService),
		SmerdLauncher:    smerd_launcher.New(nodeClients, configService),
	}
}
