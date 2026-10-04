package service_manager

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/config"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/service/secrets"
	"go.vervstack.ru/Velez/internal/service/service_manager/configurator"
	"go.vervstack.ru/Velez/internal/service/service_manager/container_manager"
	"go.vervstack.ru/Velez/internal/service/service_manager/dinds"
	"go.vervstack.ru/Velez/internal/service/service_manager/image_versions"
	"go.vervstack.ru/Velez/internal/service/service_manager/nodes_service"
	"go.vervstack.ru/Velez/internal/service/service_manager/pgaas"
	"go.vervstack.ru/Velez/internal/service/service_manager/plugins"
	"go.vervstack.ru/Velez/internal/service/service_manager/registryaas"
	"go.vervstack.ru/Velez/internal/service/service_manager/runneraas"
	"go.vervstack.ru/Velez/internal/service/service_manager/settings"
	"go.vervstack.ru/Velez/internal/service/service_manager/verv_services"
	"go.vervstack.ru/Velez/internal/service/service_manager/vervonomicon"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/local_storage"
)

type ServiceManager struct {
	containerManager *container_manager.ContainerManager
	configurator     *configurator.Configurator
	vervServices     *verv_services.VervService

	docker      node_clients.Docker
	nodeService *nodes_service.Service

	pluginService            service.PluginService
	storageContainer         *storage.Container
	secretsStore             secrets.Store
	postgresService          service.PostgresService
	runnersService           service.RunnersService
	containerRegistryService service.ContainerRegistryService
	imageVersions            service.ImageVersionsService
	settingsService          service.SettingsService
	dindService              service.DindService
}

func New(
	ctx context.Context,
	nodeClients node_clients.NodeClients,
	clusterClients cluster_clients.ClusterClients,
	cfg config.Config,
	runtimeResolver container_runtime.RuntimeResolver,
	jobsEngine jobs.Engine,
) (service.Services, error) {
	configService, err := configurator.New(clusterClients)
	if err != nil {
		return nil, rerrors.Wrap(err, "error initializing configurator")
	}

	cm := container_manager.New(nodeClients, runtimeResolver, clusterClients.StateManager())

	storageContainer := storage.NewStorageContainer(
		local_storage.New(nodeClients.Docker(), nodeClients.LocalStateManager(), cfg),
	)
	svc := plugins.New(storageContainer)
	secretsStore := secrets.New(storageContainer)

	vervServices := verv_services.New(
		clusterClients.StateManager(), cm, nodeClients.Docker(), runtimeResolver,
		vervonomicon.NewImageSource(nodeClients, runtimeResolver),
		configService,
	)

	sm := &ServiceManager{
		containerManager: cm,
		configurator:     configService,
		vervServices:     vervServices,

		docker:      nodeClients.Docker(),
		nodeService: nodes_service.New(clusterClients.StateManager()),

		pluginService:    svc,
		storageContainer: storageContainer,
		secretsStore:     secretsStore,
		// pgaas.New takes clusterClients.StateManager(), not storageContainer:
		// CreatePgInstance hands off to vervServices.CreateNewDeploy, which
		// reads/writes through clusterClients.StateManager() - the two only
		// converge onto the same Postgres-backed storage once this node joins
		// a Postgres cluster (see custom.go's InitServiceLayer), so pgaas's
		// own Services()/GetByName calls have to agree with CreateNewDeploy
		// from the start, not just after convergence.
		postgresService: pgaas.New(clusterClients.StateManager(), vervServices, secretsStore),
		// runneraas.New takes clusterClients.StateManager(), for the same
		// reason pgaas.New does just above - see that comment.
		runnersService: runneraas.New(
			clusterClients.StateManager(), vervServices, secretsStore, jobsEngine, runtimeResolver,
		),
		// registryaas.New takes clusterClients.StateManager(), for the same
		// reason pgaas.New does just above - see that comment.
		containerRegistryService: registryaas.New(clusterClients.StateManager(), vervServices, secretsStore, jobsEngine),
		imageVersions:            image_versions.New(runtimeResolver, nodeClients.Docker().Client()),
		settingsService:          settings.New(clusterClients.StateManager(), nodeClients.Docker()),
		dindService: dinds.New(
			clusterClients.StateManager(), vervServices, jobsEngine, nodeClients.Docker(),
		),
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

func (s *ServiceManager) Secrets() secrets.Store {
	return s.secretsStore
}

func (s *ServiceManager) Postgres() service.PostgresService {
	return s.postgresService
}

func (s *ServiceManager) Runners() service.RunnersService {
	return s.runnersService
}

func (s *ServiceManager) ContainerRegistry() service.ContainerRegistryService {
	return s.containerRegistryService
}

func (s *ServiceManager) ImageVersions() service.ImageVersionsService {
	return s.imageVersions
}

func (s *ServiceManager) Settings() service.SettingsService {
	return s.settingsService
}

func (s *ServiceManager) Dinds() service.DindService {
	return s.dindService
}
