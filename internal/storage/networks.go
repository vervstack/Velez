package storage

import (
	"context"

	"go.vervstack.ru/Velez/internal/domain"
)

type NetworksStorage interface {
	Create(ctx context.Context, network domain.Network) error
	Get(ctx context.Context, id string) (domain.Network, error)
	List(ctx context.Context) ([]domain.Network, error)
	Drop(ctx context.Context, id string) error
	CurrentNode(ctx context.Context) (string, error)
}

type NetworksProvider interface {
	Networks() NetworksStorage
}
