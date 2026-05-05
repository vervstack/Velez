package storage

import (
	"context"
	"database/sql"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/sqldb"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/deployments_queries"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/plugins_queries"
)

var (
	ErrAlreadyExists = rerrors.New("already exists")
	ErrNotFound      = rerrors.New("not found")
)

type Storage interface {
	Nodes() NodesStorage
	Services() ServicesStorage
	Deployments() DeploymentsStorage
	Plugins() PluginsStorage

	TxManager() *sqldb.TxManager
}

type ServicesStorage interface {
	GetById(ctx context.Context, id int64) (domain.Service, error)
	GetByName(ctx context.Context, name string) (domain.Service, error)
	UpsertService(ctx context.Context, name string) error

	List(ctx context.Context, req domain.ListServicesReq) (domain.ServiceList, error)
}

type DeploymentsStorage interface {
	List(ctx context.Context, req domain.ListDeploymentsReq) ([]domain.Deployment, error)
	ListDeployments(ctx context.Context, req domain.ListDeploymentsReq) (domain.DeploymentList, error)

	deployments_queries.Querier
	WithTx(tx *sql.Tx) *deployments_queries.Queries
}

type NodesStorage interface {
	InitNode(ctx context.Context) error
	UpdateOnline(ctx context.Context) error
	List(ctx context.Context, req domain.ListNodesReq) (domain.NodesList, error)
}

type PluginsStorage interface {
	ListPlugins(ctx context.Context) ([]domain.PluginBaseInfo, error)
	UpsertPlugin(ctx context.Context, arg plugins_queries.UpsertPluginParams) error
}
