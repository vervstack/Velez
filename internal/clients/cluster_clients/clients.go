package cluster_clients

import (
	"context"

	makosh "go.vervstack.ru/makosh/pkg/makosh_be"
	"go.vervstack.ru/matreshka/pkg/matreshka_api"

	"go.vervstack.ru/Velez/internal/domain"
)

type Configurator interface {
	matreshka_api.MatreshkaBeAPIClient
}

type ServiceDiscovery interface {
	makosh.MakoshBeAPIClient
}

type VervPrivateNetworkClient interface {
	CreateNamespace(ctx context.Context, name string) (domain.VpnNamespace, error)
	ListNamespaces(ctx context.Context) ([]domain.VpnNamespace, error)
	DeleteNamespace(ctx context.Context, id string) error

	IssueClientKey(ctx context.Context, namespace string) (string, error)
}
