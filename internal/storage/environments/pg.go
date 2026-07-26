package environments

import (
	"context"

	"go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/storage"
)

type pgStorage struct{}

func NewPg() storage.EnvironmentsStorage {
	return &pgStorage{}
}

func (p *pgStorage) ListEnvironments(_ context.Context) ([]string, error) {
	return nil, rerrors.New("not implemented")
}
