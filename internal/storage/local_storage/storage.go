package local_storage

import (
	"go.vervstack.ru/Velez/internal/clients/sqldb"
	"go.vervstack.ru/Velez/internal/storage"
)

type localStorage struct {
	nodes       *nodes
	services    *services
	deployments *deployments
	plugins     *plugins
}

func New() storage.Storage {
	return &localStorage{
		nodes:       newNodesStorage(),
		services:    newServicesStorage(),
		deployments: newDeploymentsStorage(),
		plugins:     newPluginsStorage(),
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

func (l *localStorage) TxManager() *sqldb.TxManager {
	return nil
}
