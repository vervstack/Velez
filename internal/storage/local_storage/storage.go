package local_storage

import (
	"os"

	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/sqldb"
	"go.vervstack.ru/Velez/internal/config"
	"go.vervstack.ru/Velez/internal/storage"
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

func (l *localStorage) TxManager() *sqldb.TxManager {
	return nil
}
