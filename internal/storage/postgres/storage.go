package postgres

import (
	"go.vervstack.ru/Velez/internal/clients/sqldb"
	"go.vervstack.ru/Velez/internal/storage"
	pg_queries "go.vervstack.ru/Velez/internal/storage/postgres/generated"
)

type Storage struct {
	nodeStorage *nodeStorage
}

func New(db sqldb.DB) storage.Storage {
	return &Storage{
		&nodeStorage{
			pg_queries.New(db),
		},
	}
}

func (s *Storage) Nodes() storage.NodesStorage {
	return s.nodeStorage
}
