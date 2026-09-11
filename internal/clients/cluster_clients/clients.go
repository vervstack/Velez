package cluster_clients

import (
	"context"

	"go.vervstack.ru/Velez/internal/api/clients/matreshka/pkg/matreshka_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
	makosh "go.vervstack.ru/makosh/pkg/makosh_be"
)

type ClusterClients interface {
	Configurator() Configurator
	Vpn() VervClosedNetworkClient
	ServiceDiscovery() ServiceDiscovery
	StateManager() ClusterStateManagerContainer
}

type Configurator interface {
	matreshka_api.MatreshkaApiClient
}

type ServiceDiscovery interface {
	makosh.MakoshBeAPIClient
}

type VervClosedNetworkClient interface {
	CreateNamespace(ctx context.Context, name string) (domain.VcnNamespace, error)
	GetNamespace(ctx context.Context, name string) (domain.VcnNamespace, error)
	ListNamespaces(ctx context.Context) ([]domain.VcnNamespace, error)
	DeleteNamespace(ctx context.Context, id string) error

	GetClientAuthKey(ctx context.Context, req domain.GetVcnAuthKeyReq) (domain.VcnAuthKey, error)
	IssueClientKey(ctx context.Context, req domain.IssueClientKey) (string, error)
	RegisterNode(ctx context.Context, req domain.RegisterVcnNodeReq) error
}

type ClusterStateManagerContainer interface {
	Set(state ClusterStateManager)
	ClusterStateManager
}

type ClusterStateManager interface {
	storage.Storage
}
