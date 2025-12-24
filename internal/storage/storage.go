package storage

import (
	"context"
)

type Storage interface {
	Nodes() NodesStorage
}

type NodesStorage interface {
	InitNode(ctx context.Context) error
	UpdateOnline(ctx context.Context) error
}
