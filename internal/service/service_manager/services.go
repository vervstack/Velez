package service_manager

import (
	"context"

	"github.com/sirupsen/logrus"
	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/service/service_manager/configurator"
	"go.vervstack.ru/Velez/internal/service/service_manager/container_manager"
	"go.vervstack.ru/Velez/internal/service/service_manager/nodes_service"
	"go.vervstack.ru/Velez/internal/service/service_manager/plugins"
	"go.vervstack.ru/Velez/internal/service/service_manager/verv_services"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/local_storage"
)

type ServiceManager struct {
	containerManager *container_manager.ContainerManager
	configurator     *configurator.Configurator
	vpnService       cluster_clients.VervClosedNetworkClient
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
) (service.Services, error) {

	configService, err := configurator.New(
		ctx,
		clusterClients,
		nodeClients.Docker(),
	)

	if err != nil {
		return nil, rerrors.Wrap(err, "error initializing configurator")
	}

	cm := container_manager.New(nodeClients, configService)

	storageContainer := storage.NewStorageContainer(local_storage.New(nodeClients.Docker()))
	svc := plugins.NewPluginService(storageContainer)

	sm := &ServiceManager{
		containerManager: cm,
		configurator:     configService,
		vpnService:       clusterClients.Vpn(),
		vervServices:     verv_services.New(clusterClients.StateManager(), cm, nodeClients.Docker()),

		docker:      nodeClients.Docker(),
		nodeService: nodes_service.NewService(clusterClients.StateManager()),

		pluginService:    svc,
		storageContainer: storageContainer,
	}

	// TODO VERV-128
	//go handleConfigurationSubscription(configService, sm)

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

func handleConfigurationSubscription(configurationService service.ConfigurationService, manager service.Services) {

	ctx := context.Background()

	for patch := range configurationService.GetUpdates() {
		listReq := &velez_api.ListSmerds_Request{
			Name: toolbox.ToPtr(patch.ConfigName),
		}

		smerds, err := manager.SmerdManager().ListSmerds(ctx, listReq)
		if err != nil {
			logrus.Error(rerrors.Wrap(err, "error listing smerds for configuration update hook"))
			continue
		}

		if len(smerds.Smerds) != 1 {
			logrus.Error(rerrors.New("unexpected number of smerds for configuration update hook. Expected 1 got %d", len(smerds.Smerds)))
			continue
		}

		// TODO VERV-128

	}
}
