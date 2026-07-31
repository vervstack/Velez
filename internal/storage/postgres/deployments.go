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

	serviceIDSubquery = "ds.service_id = (SELECT id FROM velez.services WHERE name = ?)"
)

type deploymentsStorage struct {
	*deployments_queries.Queries

	db sqldb.DB
}

func newDeploymentsStorage(db sqldb.DB) *deploymentsStorage {
	return &deploymentsStorage{
		Queries: deployments_queries.New(db),
		db:      db,
	}
}

func (d *deploymentsStorage) ListDeployments(ctx context.Context, req domain.ListDeploymentsReq) (
	domain.DeploymentList, error) {
	baseQuery := sq.Select().
		From("velez.deployments AS d").
		Join("velez.deployment_specifications AS ds ON d.spec_id = ds.id").
		PlaceholderFormat(sq.Dollar)

	if len(req.NodeIds) != 0 {
		baseQuery = baseQuery.Where(sq.Eq{"d.node_id": req.NodeIds})
	}

	if req.ServiceName != "" {
		baseQuery = baseQuery.Where(sq.Expr(serviceIDSubquery, req.ServiceName))
	}

	if len(req.NotStatus) != 0 {
		baseQuery = baseQuery.Where(sq.NotEq{"d.status": req.NotStatus})
	}

	total, err := countTotal(ctx, d.db, baseQuery)
	if err != nil {
		return domain.DeploymentList{}, wrapPgErr(err)
	}

	limit := req.Paging.Limit
	if limit == 0 || limit > defaultDeploymentsLimit {
		limit = defaultDeploymentsLimit
	}

	selectQuery, args, err := baseQuery.
		Columns("d.id", "ds.service_id", "d.node_id", "d.spec_id", "d.created_at", "d.updated_at", "d.status").
		Limit(limit).
		Offset(req.Paging.Offset).
		ToSql()
	if err != nil {
		return domain.DeploymentList{}, rerrors.Wrap(err, "error building sql query")
	}

	rows, err := d.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return domain.DeploymentList{}, wrapPgErr(err)
	}
	defer closeRows(rows)

	out := domain.DeploymentList{Total: total}

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
			return domain.DeploymentList{}, wrapPgErr(err)
		}

		out.Deployments = append(out.Deployments, dep)
	}

	err = rows.Err()
	if err != nil {
		return domain.DeploymentList{}, wrapPgErr(err)
	}

	return out, nil
}

func (d *deploymentsStorage) List(ctx context.Context, req domain.ListDeploymentsReq) ([]domain.Deployment, error) {
	builder := sq.Select("d.id",
		"ds.service_id",
		"d.node_id",
		"d.spec_id",
		"d.created_at",
		"d.updated_at",
		"d.status").
		From("velez.deployments AS d").
		Join("velez.deployment_specifications AS ds ON d.spec_id = ds.id")

	if len(req.NodeIds) != 0 {
		builder = builder.Where(sq.Eq{"d.node_id": req.NodeIds})
	}

	if req.ServiceName != "" {
		builder = builder.Where(sq.Expr(serviceIDSubquery, req.ServiceName))
	}

	if len(req.NotStatus) != 0 {
		builder = builder.Where(sq.NotEq{"d.status": req.NotStatus})
	}

	if req.Paging.Limit == 0 || req.Paging.Limit > defaultDeploymentsLimit {
		req.Paging.Limit = defaultDeploymentsLimit
	}

	builder = builder.
		Limit(req.Paging.Limit).
		Offset(req.Paging.Offset).
		PlaceholderFormat(sq.Dollar)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, rerrors.Wrap(err, "error building sql query")
	}

	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, wrapPgErr(err)
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
			return nil, wrapPgErr(err)
		}

		out = append(out, dep)
	}

	err = rows.Err()
	if err != nil {
		return nil, wrapPgErr(err)
	}

	return out, nil
}
