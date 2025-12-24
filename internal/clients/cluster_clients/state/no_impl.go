package state

import (
	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/storage"
)

type noImpl struct {
}

var _ cluster_clients.ClusterStateManager = (*noImpl)(nil)

func (n noImpl) Nodes() storage.NodesStorage {
	return nil
}
