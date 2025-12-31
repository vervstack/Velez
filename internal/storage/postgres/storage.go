package postgres

import (
	"go.vervstack.ru/Velez/internal/clients/sqldb"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/nodes_queries"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/services_queries"
)

type Storage struct {
	nodeStorage     *nodeStorage
	servicesStorage *servicesStorage
}

func New(db sqldb.DB) storage.Storage {
	return &Storage{
		&nodeStorage{
			nodes_queries.New(db),
		},
		&servicesStorage{
			services_queries.New(db),
		},
	}
}

func (s *Storage) Nodes() storage.NodesStorage {
	return s.nodeStorage
}

func (s *Storage) Services() storage.ServicesStorage {
	return s.servicesStorage
}
