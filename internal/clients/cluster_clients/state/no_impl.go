package state

import (
	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
)

type noImpl struct {
}

var _ cluster_clients.ClusterStateManager = (*noImpl)(nil)
