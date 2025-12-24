package state

import (
	"sync/atomic"

	"github.com/rs/zerolog/log"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/state"
	"go.vervstack.ru/Velez/internal/storage"
)

type stateManager struct {
	state atomic.Pointer[cluster_clients.ClusterStateManager]
}

func New(state state.State) (cluster_clients.ClusterStateManagerContainer, error) {
	sm := &stateManager{
		state: atomic.Pointer[cluster_clients.ClusterStateManager]{},
	}

	sm.Set(&noImpl{})

	if state.PgRootDsn != "" {
		// TODO implement strict and not strict mode. In non-strict when unable to connect to pg - not fail
		pg, err := NewPgStateManager(state.PgRootDsn)
		if err != nil {
			log.Err(err).
				Msg("error connecting to PgRootDsn")
		} else {
			sm.Set(pg)
		}
	}

	return sm, nil
}

func (s *stateManager) Nodes() storage.NodesStorage {
	l := s.state.Load()
	cm, ok := (*l).(cluster_clients.ClusterStateManager)
	if !ok {
		return nil
	}

	return cm.Nodes()
}

func (s *stateManager) Set(manager cluster_clients.ClusterStateManager) {
	s.state.Store(&manager)
}
