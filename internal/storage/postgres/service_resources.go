package postgres

import (
	"context"

	"go.vervstack.ru/Velez/internal/clients/sqldb"
	"go.vervstack.ru/Velez/internal/domain"
	service_resources_queries "go.vervstack.ru/Velez/internal/storage/postgres/generated/service_resources_queries"
)

type serviceResourcesStorage struct {
	*service_resources_queries.Queries
}

func newServiceResourcesStorage(db sqldb.DB) *serviceResourcesStorage {
	return &serviceResourcesStorage{
		Queries: service_resources_queries.New(db),
	}
}

func (s *serviceResourcesStorage) GetResources(ctx context.Context, serviceName string) ([]domain.BoundResource, error) {
	rows, err := s.Queries.GetServiceResources(ctx, serviceName)
	if err != nil {
		return nil, wrapPgErr(err)
	}

	result := make([]domain.BoundResource, len(rows))
	for i, row := range rows {
		result[i] = domain.BoundResource{
			Name:         row.ResourceName,
			ResourceType: row.ResourceType,
			Status:       row.Status,
		}
	}

	return result, nil
}

func (s *serviceResourcesStorage) UpsertResource(ctx context.Context, serviceName, resourceName, resourceType string) error {
	params := service_resources_queries.UpsertServiceResourceParams{
		ServiceName:  serviceName,
		ResourceName: resourceName,
		ResourceType: resourceType,
	}
	err := s.Queries.UpsertServiceResource(ctx, params)
	if err != nil {
		return wrapPgErr(err)
	}
	return nil
}
