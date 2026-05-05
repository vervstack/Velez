package state

import (
	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/clients/sqldb"
	"go.vervstack.ru/Velez/internal/storage"
)

type localState struct {
}

func New() cluster_clients.ClusterStateManager {
	return &localState{}
}

func (l localState) Nodes() storage.NodesStorage {

}

func (l localState) Services() storage.ServicesStorage {

}

func (l localState) Deployments() storage.DeploymentsStorage {

}

func (l localState) Plugins() storage.PluginsStorage {

}

func (l localState) TxManager() *sqldb.TxManager {
	//TODO
	return nil
}
