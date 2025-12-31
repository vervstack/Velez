package storage

import (
	"context"

	"go.vervstack.ru/Velez/internal/storage/postgres/generated/services_queries"
)

type Storage interface {
	Nodes() NodesStorage
	Services() ServicesStorage
}

type NodesStorage interface {
	InitNode(ctx context.Context) error
	UpdateOnline(ctx context.Context) error
}

type ServicesStorage interface {
	services_queries.Querier
}
