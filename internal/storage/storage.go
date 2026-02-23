package storage

import (
	"context"
	"database/sql"

	"go.vervstack.ru/Velez/internal/clients/sqldb"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/deployments_queries"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/services_queries"
)

type Storage interface {
	Nodes() NodesStorage
	Services() ServicesStorage
	Deployments() DeploymentsStorage

	TxManager() *sqldb.TxManager
}

type NodesStorage interface {
	InitNode(ctx context.Context) error
	UpdateOnline(ctx context.Context) error
}

type ServicesStorage interface {
	services_queries.Querier

	List(ctx context.Context, req domain.ListServicesReq) (domain.ServiceList, error)
}

type DeploymentsStorage interface {
	List(ctx context.Context, req domain.ListDeploymentsReq) ([]domain.Deployment, error)

	deployments_queries.Querier
	WithTx(tx *sql.Tx) *deployments_queries.Queries
}
