package cluster_steps

import (
	"context"

	"go.vervstack.ru/Velez/internal/pipelines/steps"
)

type storeDatabaseConnection struct {
}

func StoreDatabaseConnection() steps.Step {
	return &storeDatabaseConnection{}
}

func (s *storeDatabaseConnection) Do(ctx context.Context) error {
	return nil
}
