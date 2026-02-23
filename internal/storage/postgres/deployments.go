package postgres

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/sqldb"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/deployments_queries"
)

const (
	defaultDeploymentsLimit = 100
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
	q := sq.Select("id",
		"service_id",
		"node_id",
		"spec_id",
		"created_at",
		"updated_at",
		"status").
		From("velez.deployments")

	if len(req.NodeIds) != 0 {
		q = q.Where(sq.Eq{"node_id": req.NodeIds})
	}

	if len(req.ServiceIds) != 0 {
		q = q.Where(sq.Eq{"service_id": req.ServiceIds})
	}

	if len(req.NotStatus) != 0 {
		q = q.Where(sq.NotEq{"status": req.NotStatus})
	}

	if req.Paging.Limit == 0 || req.Paging.Limit > defaultDeploymentsLimit {
		req.Paging.Limit = defaultDeploymentsLimit
	}
	q = q.
		Limit(req.Paging.Limit).
		Offset(req.Paging.Offset).
		PlaceholderFormat(sq.Dollar)

	query, args, err := q.ToSql()
	if err != nil {
		return nil, rerrors.Wrap(err, "error building sql query")
	}

	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, rerrors.Wrap(err, "error quering")
	}
	defer closeRows(rows)

	out := make([]domain.Deployment, 0, req.Paging.Limit)

	for rows.Next() {
		var dep domain.Deployment
		err = rows.Scan(
			&dep.Id,
			&dep.ServiceId,
			&dep.NodeId,
			&dep.SpecId,
			&dep.CreatedAt,
			&dep.UpdatedAt,
			&dep.Status,
		)
		if err != nil {
			return nil, rerrors.Wrap(err, "error scanning row")
		}

		out = append(out, dep)
	}

	return out, nil
}
