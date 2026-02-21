package state

import (
	"sync/atomic"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/clients/sqldb"
	"go.vervstack.ru/Velez/internal/storage"
)

type stateManager struct {
	state atomic.Pointer[cluster_clients.ClusterStateManager]
}

func NewContainer() cluster_clients.ClusterStateManagerContainer {
	sm := &stateManager{
		state: atomic.Pointer[cluster_clients.ClusterStateManager]{},
	}

	sm.Set(&noImpl{})

	return sm
}

func (s *stateManager) Set(manager cluster_clients.ClusterStateManager) {
	s.state.Store(&manager)
}

func (s *stateManager) Nodes() storage.NodesStorage {
	l := s.state.Load()
	cm, ok := (*l).(cluster_clients.ClusterStateManager)
	if !ok {
		return nil
	}

	return cm.Nodes()
}

func (s *stateManager) Services() storage.ServicesStorage {
	l := s.state.Load()
	cm, ok := (*l).(cluster_clients.ClusterStateManager)
	if !ok {
		return nil
	}

	return cm.Services()
}

func (s *stateManager) Deployments() storage.DeploymentsStorage {
	l := s.state.Load()
	cm, ok := (*l).(cluster_clients.ClusterStateManager)
	if !ok {
		return nil
	}

	return cm.Deployments()
}

func (s *stateManager) TxManager() *sqldb.TxManager {
	l := s.state.Load()
	cm, ok := (*l).(cluster_clients.ClusterStateManager)
	if !ok {
		return nil
	}

	return cm.TxManager()
}

func (s *stateManager) tryConnect() {

}
