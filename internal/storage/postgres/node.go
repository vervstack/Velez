package postgres

import (
	"context"

	"go.redsock.ru/rerrors"

	pg_queries "go.vervstack.ru/Velez/internal/storage/postgres/generated"
)

type nodeStorage struct {
	querier pg_queries.Querier
}

func (n *nodeStorage) InitNode(ctx context.Context) error {
	_, err := n.querier.InitNode(ctx)
	if err != nil {
		return rerrors.Wrap(err, "error initializing new node")
	}

	return nil
}

func (n *nodeStorage) UpdateOnline(ctx context.Context) error {
	err := n.querier.UpdateOnline(ctx)
	if err != nil {
		return rerrors.Wrap(err, "error updating node's last online")
	}

	return nil
}
