package postgres

import (
	"context"

	"go.vervstack.ru/Velez/internal/clients/sqldb"
)

type userStorage struct {
	db sqldb.DB
}

func (u *userStorage) CreateUser(ctx context.Context) {

}
