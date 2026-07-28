package service_manager

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/config"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/service/service_manager/configurator"
	"go.vervstack.ru/Velez/internal/service/service_manager/container_manager"
	"go.vervstack.ru/Velez/internal/service/service_manager/nodes_service"
	"go.vervstack.ru/Velez/internal/service/service_manager/plugins"
	"go.vervstack.ru/Velez/internal/service/service_manager/verv_services"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/environments"
	"go.vervstack.ru/Velez/internal/storage/local_storage"
)

type ServiceManager struct {
	containerManager *container_manager.ContainerManager
	configurator     *configurator.Configurator
	vervServices     *verv_services.VervService

	docker      node_clients.Docker
	nodeService *nodes_service.Service

	pluginService    service.PluginService
	storageContainer *storage.Container
}

func New(
	ctx context.Context,
	nodeClients node_clients.NodeClients,
	clusterClients cluster_clients.ClusterClients,
	envs []string,
	cfg config.Config,
) (service.Services, error) {
	configService, err := configurator.New(clusterClients)
	if err != nil {
		return nil, rerrors.Wrap(err, "error initializing configurator")
	}

	cm := container_manager.New(nodeClients)

	storageContainer := storage.NewStorageContainer(local_storage.New(nodeClients.Docker(), cfg))
	svc := plugins.NewPluginService(storageContainer)

	envStorageContainer := environments.NewContainer(environments.NewStatic(envs))

	sm := &ServiceManager{
		containerManager: cm,
		configurator:     configService,
		vervServices:     verv_services.New(clusterClients.StateManager(), cm, nodeClients.Docker(), envStorageContainer),

		docker:      nodeClients.Docker(),
		nodeService: nodes_service.NewService(clusterClients.StateManager()),

		pluginService:    svc,
		storageContainer: storageContainer,
	}

	// TODO VERV-128
	// go handleConfigurationSubscription(configService, sm)

	return sm, nil
}

func (s *ServiceManager) VervServices() service.VervServicesService {
	return s.vervServices
}

func (s *ServiceManager) SmerdManager() service.ContainerService {
	return s.containerManager
}

func (s *ServiceManager) ConfigurationService() service.ConfigurationService {
	return s.configurator
}

func (s *ServiceManager) Docker() node_clients.Docker {
	return s.docker
}

func (s *ServiceManager) NodeService() service.NodeService {
	return s.nodeService
}

func (s *ServiceManager) PluginService() service.PluginService {
	return s.pluginService
}

func (s *ServiceManager) StorageContainer() *storage.Container {
	return s.storageContainer
}
