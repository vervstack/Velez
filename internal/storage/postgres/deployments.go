package postgres

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/sqldb"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/deployments_queries"
)

type deploymentsStorage struct {
	db sqldb.DB

	*deployments_queries.Queries
}

func newDeploymentsStorage(db sqldb.DB) *deploymentsStorage {
	return &deploymentsStorage{
		db:      db,
		Queries: deployments_queries.New(db),
	}
}

func (d *deploymentsStorage) List(ctx context.Context, req domain.ListDeploymentsReq) ([]domain.Deployment, error) {
	return nil, rerrors.New("unimplemented")
}
