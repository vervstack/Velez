package local_storage

import (
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/sqldb"
	"go.vervstack.ru/Velez/internal/storage"
)

type localStorage struct {
	nodes            *nodes
	services         *services
	deployments      *deployments
	plugins          *dockerPluginsStorage
	serviceDeps      *dockerServiceDepsStorage
	serviceResources *dockerServiceResourcesStorage
}

func New(containerApi node_clients.Docker) storage.Storage {
	return &localStorage{
		nodes:            newNodesStorage(),
		services:         newServicesStorage(),
		deployments:      newDeploymentsStorage(),
		plugins:          newPluginsStorage(containerApi),
		serviceDeps:      newServiceDepsStorage(containerApi),
		serviceResources: newServiceResourcesStorage(containerApi),
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

func (l *localStorage) TxManager() *sqldb.TxManager {
	return nil
}
