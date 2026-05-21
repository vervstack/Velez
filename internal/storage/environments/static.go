package environments

import (
	"context"

	"go.vervstack.ru/Velez/internal/storage"
)

type staticStorage struct {
	envs []string
}

func NewStatic(envs []string) storage.EnvironmentsStorage {
	return &staticStorage{envs: envs}
}

func (s *staticStorage) ListEnvironments(_ context.Context) ([]string, error) {
	return s.envs, nil
}
