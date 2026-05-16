package postgres

import (
	"context"

	"go.vervstack.ru/Velez/internal/clients/sqldb"
	"go.vervstack.ru/Velez/internal/domain"
	service_dependencies_queries "go.vervstack.ru/Velez/internal/storage/postgres/generated/service_dependencies_queries"
)

type serviceDependenciesStorage struct {
	*service_dependencies_queries.Queries
}

func newServiceDependenciesStorage(db sqldb.DB) *serviceDependenciesStorage {
	return &serviceDependenciesStorage{
		Queries: service_dependencies_queries.New(db),
	}
}

func (s *serviceDependenciesStorage) UpsertDependency(ctx context.Context, source, target, proto string) error {
	params := service_dependencies_queries.UpsertServiceDependencyParams{
		SourceService: source,
		TargetService: target,
		Proto:         proto,
	}
	err := s.Queries.UpsertServiceDependency(ctx, params)
	if err != nil {
		return wrapPgErr(err)
	}
	return nil
}

func (s *serviceDependenciesStorage) GetDependencies(ctx context.Context, serviceName string) ([]domain.ServiceDependency, error) {
	rows, err := s.Queries.GetServiceDependencies(ctx, serviceName)
	if err != nil {
		return nil, wrapPgErr(err)
	}

	result := make([]domain.ServiceDependency, len(rows))
	for i, row := range rows {
		result[i] = domain.ServiceDependency{
			SourceService: serviceName,
			TargetService: row.TargetService,
			Proto:         row.Proto,
		}
	}

	return result, nil
}

func (s *serviceDependenciesStorage) GetCallers(ctx context.Context, serviceName string) ([]domain.ServiceDependency, error) {
	rows, err := s.Queries.GetServiceCallers(ctx, serviceName)
	if err != nil {
		return nil, wrapPgErr(err)
	}

	result := make([]domain.ServiceDependency, len(rows))
	for i, row := range rows {
		result[i] = domain.ServiceDependency{
			SourceService: row.SourceService,
			TargetService: serviceName,
			Proto:         row.Proto,
		}
	}

	return result, nil
}
