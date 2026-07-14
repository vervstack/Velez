package local_storage

import (
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/sqldb"
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

func New(containerApi node_clients.Docker) storage.Storage {
	return &localStorage{
		nodes:            newNodesStorage(),
		services:         newServicesStorage(containerApi),
		deployments:      newDeploymentsStorage(),
		plugins:          newPluginsStorage(containerApi),
		serviceDeps:      newServiceDepsStorage(containerApi),
		serviceResources: newServiceResourcesStorage(containerApi),
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
