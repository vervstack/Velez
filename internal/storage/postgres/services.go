package postgres

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/sqldb"
	"go.vervstack.ru/Velez/internal/domain"
	pg_queries "go.vervstack.ru/Velez/internal/storage/postgres/generated/services_queries"
)

type servicesStorage struct {
	conn sqldb.DB
	pg_queries.Querier
}

var listServiceHelper = serviceBaseInfoHelper{}

func (s *servicesStorage) List(ctx context.Context, req domain.ListServicesReq) (domain.ServiceList, error) {
	baseQuery := listServiceHelper.buildListQuery(req)

	totalRows, err := countTotal(ctx, s.conn, baseQuery)
	if err != nil {
		return domain.ServiceList{}, wrapPgErr(err)
	}

	baseQuery = baseQuery.Limit(min(req.Paging.Limit, totalRows))

	selectQuery, args, err := baseQuery.Columns(listServiceHelper.columns()...).ToSql()
	if err != nil {
		return domain.ServiceList{}, rerrors.Wrap(err, "error building list query")
	}

	rows, err := s.conn.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return domain.ServiceList{}, wrapPgErr(err)
	}

	defer closeRows(rows)

	out := domain.ServiceList{}

	for rows.Next() {
		var serviceBaseInfo domain.ServiceBaseInfo
		serviceBaseInfo, err = listServiceHelper.scanServiceBaseInfo(rows)
		if err != nil {
			return domain.ServiceList{}, wrapPgErr(err)
		}

		out.Services = append(out.Services, serviceBaseInfo)
	}

	return out, nil
}

type serviceBaseInfoHelper struct {
}

func (s serviceBaseInfoHelper) buildListQuery(req domain.ListServicesReq) sq.SelectBuilder {
	q := sq.Select().
		From("velez.services").
		PlaceholderFormat(sq.Dollar)

	if req.NamePattern.Valid {
		q = q.Where(sq.ILike{
			"name": req.NamePattern.Value,
		})
	}

	return q
}

func (s serviceBaseInfoHelper) columns() []string {
	return []string{"id", "name"}
}

func (s serviceBaseInfoHelper) scanServiceBaseInfo(row sqldb.Scannable) (baseInfo domain.ServiceBaseInfo, err error) {
	err = row.Scan(
		&baseInfo.Id,
		&baseInfo.Name,
	)
	if err != nil {
		return baseInfo, rerrors.Wrap(err)
	}

	return baseInfo, nil
}

func countTotal(ctx context.Context, conn sqldb.DB, baseQuery sq.SelectBuilder) (uint64, error) {
	var totalRows uint64
	countQuery, args, err := baseQuery.Columns("count(*)").ToSql()
	if err != nil {
		return totalRows, rerrors.Wrap(err, "building count query")
	}

	err = conn.QueryRowContext(ctx, countQuery, args...).
		Scan(&totalRows)
	if err != nil {
		return totalRows, rerrors.Wrap(err, "scanning count query")
	}

	return totalRows, nil
}
