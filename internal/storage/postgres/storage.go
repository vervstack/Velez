package postgres

import (
	"go.vervstack.ru/Velez/internal/clients/sqldb"
	"go.vervstack.ru/Velez/internal/storage"
)

type Storage struct {
	userStorage *userStorage
}

func New(db sqldb.DB) storage.Storage {
	return &Storage{
		userStorage: &userStorage{db},
	}
}

func (s *Storage) User() storage.Users {
	return s.userStorage
}
