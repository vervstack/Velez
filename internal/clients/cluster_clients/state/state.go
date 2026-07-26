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

func NewContainer(init cluster_clients.ClusterStateManager) cluster_clients.ClusterStateManagerContainer {
	sm := &stateManager{
		state: atomic.Pointer[cluster_clients.ClusterStateManager]{},
	}

	sm.Set(init)

	return sm
}

func (s *stateManager) Set(manager cluster_clients.ClusterStateManager) {
	s.state.Store(&manager)
}

func (s *stateManager) Nodes() storage.NodesStorage {
	l := s.state.Load()
	if l == nil {
		return nil
	}

	return (*l).Nodes()
}

func (s *stateManager) Services() storage.ServicesStorage {
	l := s.state.Load()
	if l == nil {
		return nil
	}

	return (*l).Services()
}

func (s *stateManager) Deployments() storage.DeploymentsStorage {
	l := s.state.Load()
	if l == nil {
		return nil
	}

	return (*l).Deployments()
}

func (s *stateManager) TxManager() *sqldb.TxManager {
	l := s.state.Load()
	if l == nil {
		return nil
	}

	return (*l).TxManager()
}

func (s *stateManager) Plugins() storage.PluginsStorage {
	l := s.state.Load()
	if l == nil {
		return nil
	}

	return (*l).Plugins()
}

func (s *stateManager) ServiceDependencies() storage.ServiceDependenciesStorage {
	l := s.state.Load()
	if l == nil {
		return nil
	}

	return (*l).ServiceDependencies()
}

func (s *stateManager) ServiceResources() storage.ServiceResourcesStorage {
	l := s.state.Load()
	if l == nil {
		return nil
	}

	return (*l).ServiceResources()
}

func (s *stateManager) Tasks() storage.TasksStorage {
	l := s.state.Load()
	if l == nil {
		return nil
	}

	return (*l).Tasks()
}

func (s *stateManager) Jobs() storage.JobsStorage {
	l := s.state.Load()
	if l == nil {
		return nil
	}

	return (*l).Jobs()
}
