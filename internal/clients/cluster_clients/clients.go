package cluster_clients

import (
	"context"

	"go.redsock.ru/rerrors"
	makosh "go.vervstack.ru/makosh/pkg/makosh_be"
	"go.vervstack.ru/matreshka/pkg/matreshka_api"
	"google.golang.org/grpc/codes"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
)

var (
	ErrServiceIsDisabled = rerrors.New("service is disabled", codes.FailedPrecondition)
)

type ClusterClients interface {
	Configurator() Configurator
	Vpn() VervClosedNetworkClient
	ServiceDiscovery() ServiceDiscovery
	StateManager() ClusterStateManagerContainer
}

type Configurator interface {
	matreshka_api.MatreshkaBeAPIClient
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
