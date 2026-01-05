package state

import (
	"context"
	"database/sql"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/clients/sqldb"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/deployments_queries"
)

type noImpl struct {
}

var _ cluster_clients.ClusterStateManager = (*noImpl)(nil)

func (n noImpl) Deployments() storage.DeploymentsStorage {
	return &noImplPg{}
}

func (n noImpl) TxManager() *sqldb.TxManager {
	return nil
}

func (n noImpl) Services() storage.ServicesStorage {
	return nil
}

func (n noImpl) Nodes() storage.NodesStorage {
	return nil
}

type noImplPg struct {
}

func (n *noImplPg) List(ctx context.Context, req domain.ListDeploymentsReq) ([]domain.Deployment, error) {
	return nil, cluster_clients.ErrServiceIsDisabled
}

func (n *noImplPg) CreateDeployment(ctx context.Context, arg deployments_queries.CreateDeploymentParams) (interface{}, error) {
	return nil, cluster_clients.ErrServiceIsDisabled
}

func (n *noImplPg) CreateSpecification(ctx context.Context, arg deployments_queries.CreateSpecificationParams) (int64, error) {
	return 0, cluster_clients.ErrServiceIsDisabled
}

func (n *noImplPg) GetSpecificationById(ctx context.Context, id int64) (deployments_queries.VelezDeploymentSpecification, error) {
	return deployments_queries.VelezDeploymentSpecification{}, cluster_clients.ErrServiceIsDisabled
}

func (n *noImplPg) UpdateDeploymentStatus(ctx context.Context, arg deployments_queries.UpdateDeploymentStatusParams) error {
	return cluster_clients.ErrServiceIsDisabled
}

func (n *noImplPg) WithTx(tx *sql.Tx) *deployments_queries.Queries {
	q := deployments_queries.Queries{}
	return q.WithTx(tx)
}
