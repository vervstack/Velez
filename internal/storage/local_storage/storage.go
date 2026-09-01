package local_storage

import (
	"os"

	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/sqldb"
	"go.vervstack.ru/Velez/internal/config"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/environments"
	"go.vervstack.ru/Velez/internal/storage/registries"
)

// allEnvironments is the empty Docker.ListContainers suffix - these
// docker-backed views are node-wide (every service/dependency/resource on this
// node, regardless of which environment owns it), so they deliberately don't
// scope by labels.SuffixLabel.
const (
	allEnvironments = ""
)

type localStorage struct {
	nodes            *nodes
	services         *dockerServices
	deployments      *deployments
	plugins          *dockerPluginsStorage
	serviceDeps      *dockerServiceDepsStorage
	serviceResources *dockerServiceResourcesStorage
	tasks            *tasks
	jobs             *jobs
	environments     storage.EnvironmentsStorage
	registries       storage.RegistriesStorage
}

func New(containerAPI node_clients.Docker, cfg config.Config) storage.Storage {
	region := cfg.Environment.NodeRegion
	if region == "" {
		hostname, err := os.Hostname()
		if err == nil {
			region = hostname
		}
	}

	return &localStorage{
		nodes:            newNodesStorage(containerAPI, region),
		services:         newServicesStorage(containerAPI),
		deployments:      newDeploymentsStorage(),
		plugins:          newPluginsStorage(containerAPI),
		serviceDeps:      newServiceDepsStorage(containerAPI),
		serviceResources: newServiceResourcesStorage(containerAPI),
		tasks:            newTasksStorage(),
		jobs:             newJobsStorage(),
		// Single-node/dev mode has no velez.environments table, so the
		// in-memory storage is seeded from config to mirror what the
		// migrations seed in cluster mode.
		environments: environments.NewStatic(cfg.Environment.Environments, cfg.Environment.ContainerSuffix),
		// Single-node/dev mode has no velez.registries table either - see
		// registries.NewStatic.
		registries: registries.NewStatic(),
	}
}

func (l *localStorage) Nodes() storage.NodesStorage {
	return l.nodes
}

func (l *localStorage) Services() storage.ServicesStorage {
	return l.services
}

func (l *localStorage) Deployments() storage.DeploymentsStorage {
	return l.deployments
}

func (l *localStorage) Plugins() storage.PluginsStorage {
	return l.plugins
}

func (l *localStorage) ServiceDependencies() storage.ServiceDependenciesStorage {
	return l.serviceDeps
}

func (l *localStorage) ServiceResources() storage.ServiceResourcesStorage {
	return l.serviceResources
}

func (l *localStorage) Tasks() storage.TasksStorage {
	return l.tasks
}

func (l *localStorage) Jobs() storage.JobsStorage {
	return l.jobs
}

func (l *localStorage) Environments() storage.EnvironmentsStorage {
	return l.environments
}

func (l *localStorage) Registries() storage.RegistriesStorage {
	return l.registries
}

func (l *localStorage) TxManager() *sqldb.TxManager {
	return nil
}
