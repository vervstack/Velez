package state

import (
	"context"
	"database/sql"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/clients/sqldb"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/deployments_queries"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/plugins_queries"
)

type noImpl struct{}

var _ cluster_clients.ClusterStateManager = (*noImpl)(nil)

func (n noImpl) Deployments() storage.DeploymentsStorage {
	return &noImplDeployments{}
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

func (n noImpl) Plugins() storage.PluginsStorage {
	return &noImplPlugins{}
}

type noImplDeployments struct{}

func (n *noImplDeployments) List(_ context.Context, _ domain.ListDeploymentsReq) ([]domain.Deployment, error) {
	return nil, cluster_clients.ErrServiceIsDisabled
}

func (n *noImplDeployments) ListDeployments(_ context.Context, _ domain.ListDeploymentsReq) (domain.DeploymentList, error) {
	return domain.DeploymentList{}, cluster_clients.ErrServiceIsDisabled
}

func (n *noImplDeployments) CreateDeployment(_ context.Context, _ deployments_queries.CreateDeploymentParams) (interface{}, error) {
	return nil, cluster_clients.ErrServiceIsDisabled
}

func (n *noImplDeployments) CreateSpecification(_ context.Context, _ deployments_queries.CreateSpecificationParams) (int64, error) {
	return 0, cluster_clients.ErrServiceIsDisabled
}

func (n *noImplDeployments) GetSpecificationById(_ context.Context, _ int64) (deployments_queries.GetSpecificationByIdRow, error) {
	return deployments_queries.GetSpecificationByIdRow{}, cluster_clients.ErrServiceIsDisabled
}

func (n *noImplDeployments) UpdateDeploymentStatus(_ context.Context, _ deployments_queries.UpdateDeploymentStatusParams) error {
	return cluster_clients.ErrServiceIsDisabled
}

func (n *noImplDeployments) WithTx(tx *sql.Tx) *deployments_queries.Queries {
	q := deployments_queries.Queries{}
	return q.WithTx(tx)
}

type noImplPlugins struct{}

func (n *noImplPlugins) ListPlugins(_ context.Context) ([]domain.PluginBaseInfo, error) {
	return nil, cluster_clients.ErrServiceIsDisabled
}

func (n *noImplPlugins) UpsertPlugin(_ context.Context, _ plugins_queries.UpsertPluginParams) error {
	return cluster_clients.ErrServiceIsDisabled
}
