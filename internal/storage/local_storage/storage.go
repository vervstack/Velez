package local_storage

import (
	"go.vervstack.ru/Velez/internal/clients/sqldb"
	"go.vervstack.ru/Velez/internal/storage"
)

type localStorage struct {
	nodes *nodes
}

func New() storage.Storage {
	return &localStorage{
		newNodesStorage(),
	}
}

func (l *localStorage) Nodes() storage.NodesStorage {
	return l.nodes
}

func (l *localStorage) Services() storage.ServicesStorage {
	return nil // TODO
}

func (l *localStorage) Deployments() storage.DeploymentsStorage {
	return nil // TODO
}

func (l *localStorage) Plugins() storage.PluginsStorage {
	return nil // TODO
}

func (l *localStorage) TxManager() *sqldb.TxManager {
	return nil
}
