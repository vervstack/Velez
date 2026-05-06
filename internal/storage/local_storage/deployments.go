package local_storage

import (
	"context"
	"database/sql"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/deployments_queries"
)

type deployments struct{}

func newDeploymentsStorage() *deployments {
	return &deployments{}
}

func (d *deployments) List(_ context.Context, _ domain.ListDeploymentsReq) ([]domain.Deployment, error) {
	return nil, nil
}

func (d *deployments) ListDeployments(_ context.Context, _ domain.ListDeploymentsReq) (domain.DeploymentList, error) {
	return domain.DeploymentList{}, nil
}

func (d *deployments) CreateDeployment(_ context.Context, _ deployments_queries.CreateDeploymentParams) (interface{}, error) {
	return nil, nil
}

func (d *deployments) CreateSpecification(_ context.Context, _ deployments_queries.CreateSpecificationParams) (int64, error) {
	return 0, nil
}

func (d *deployments) GetSpecificationById(_ context.Context, _ int64) (deployments_queries.GetSpecificationByIdRow, error) {
	return deployments_queries.GetSpecificationByIdRow{}, nil
}

func (d *deployments) UpdateDeploymentStatus(_ context.Context, _ deployments_queries.UpdateDeploymentStatusParams) error {
	return nil
}

func (d *deployments) WithTx(_ *sql.Tx) *deployments_queries.Queries {
	return nil
}
