package postgres

import (
	"context"
	"sort"

	sq "github.com/Masterminds/squirrel"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/sqldb"
	"go.vervstack.ru/Velez/internal/domain"
	service_resources_queries "go.vervstack.ru/Velez/internal/storage/postgres/generated/service_resources_queries"
	pg_queries "go.vervstack.ru/Velez/internal/storage/postgres/generated/services_queries"
)

const (
	serviceNameColumn = "s.name"
)

type servicesStorage struct {
	conn            sqldb.DB
	querier         pg_queries.Querier
	resourceQuerier service_resources_queries.Querier
}

func (s *servicesStorage) GetByName(ctx context.Context, name string) (domain.Service, error) {
	row, err := s.querier.GetByName(ctx, name)
	if err != nil {
		return domain.Service{}, wrapPgErr(err)
	}

	return fromStorageToDomainService(row), nil
}

func (s *servicesStorage) UpsertService(ctx context.Context, name string) error {
	err := s.querier.UpsertService(ctx, name)
	if err != nil {
		return wrapPgErr(err)
	}

	return nil
}

func (s *servicesStorage) Delete(ctx context.Context, name string) error {
	err := s.querier.DeleteByName(ctx, name)
	if err != nil {
		return wrapPgErr(err)
	}

	return nil
}

func fromStorageToDomainService(row pg_queries.VelezService) domain.Service {
	return domain.Service{
		ID: row.ID,
		ServiceBaseInfo: domain.ServiceBaseInfo{
			Name: row.Name,
		},
	}
}

var listServiceHelper = serviceBaseInfoHelper{}

func (s *servicesStorage) List(ctx context.Context, req domain.ListServicesReq) (domain.ServiceList, error) {
	resourceRows, err := s.resourceQuerier.ListDistinctResourceNames(ctx)
	if err != nil {
		return domain.ServiceList{}, wrapPgErr(err)
	}

	resourceTypeByName := make(map[string]string, len(resourceRows))
	resourceNames := make([]string, 0, len(resourceRows))

	for _, row := range resourceRows {
		resourceTypeByName[row.ResourceName] = row.ResourceType
		resourceNames = append(resourceNames, row.ResourceName)
	}

	baseQuery := listServiceHelper.buildListQuery(req, resourceNames)

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

	out := domain.ServiceList{Total: totalRows}

	for rows.Next() {
		var serviceBaseInfo domain.ServiceBaseInfo

		serviceBaseInfo, err = listServiceHelper.scanServiceBaseInfo(rows)
		if err != nil {
			return domain.ServiceList{}, wrapPgErr(err)
		}

		serviceBaseInfo.Labels = domain.ClassifyService(serviceBaseInfo.Name, resourceTypeByName[serviceBaseInfo.Name])

		out.Services = append(out.Services, serviceBaseInfo)
	}

	err = rows.Err()
	if err != nil {
		return domain.ServiceList{}, wrapPgErr(err)
	}

	return out, nil
}

type serviceBaseInfoHelper struct{}

// buildListQuery builds the dynamic service-list query. resourceNames is the
// distinct set of velez.service_resources.resource_name values; when
// req.IncludeInternal is false, rows whose name is a core service or a bound
// resource are filtered out in SQL so countTotal stays consistent with the
// returned page.
func (s serviceBaseInfoHelper) buildListQuery(req domain.ListServicesReq, resourceNames []string) sq.SelectBuilder {
	query := sq.Select().
		From("velez.services s").
		LeftJoin("(SELECT ds.service_id, MAX(d.created_at) AS last_deployed_at FROM velez.deployments d " +
			"JOIN velez.deployment_specifications ds ON ds.id = d.spec_id GROUP BY ds.service_id) " +
			"ld ON ld.service_id = s.id").
		PlaceholderFormat(sq.Dollar)

	if req.NamePattern.Valid {
		query = query.Where(sq.ILike{
			serviceNameColumn: req.NamePattern.Value,
		})
	}

	if !req.IncludeInternal {
		query = query.Where(sq.NotEq{serviceNameColumn: coreServiceNamesList()})

		if len(resourceNames) > 0 {
			query = query.Where(sq.NotEq{serviceNameColumn: resourceNames})
		}
	}

	return query
}

// coreServiceNamesList returns domain.CoreServiceNames as a sorted slice so the
// generated SQL is deterministic.
func coreServiceNamesList() []string {
	names := make([]string, 0, len(domain.CoreServiceNames))
	for name := range domain.CoreServiceNames {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}

func (s serviceBaseInfoHelper) columns() []string {
	return []string{serviceNameColumn, "ld.last_deployed_at"}
}

func (s serviceBaseInfoHelper) scanServiceBaseInfo(row sqldb.Scannable) (baseInfo domain.ServiceBaseInfo, err error) {
	err = row.Scan(
		&baseInfo.Name,
		&baseInfo.LastDeployedAt,
	)
	if err != nil {
		return baseInfo, rerrors.Wrap(err, "error scanning service base info")
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
