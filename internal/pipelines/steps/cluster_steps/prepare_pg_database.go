package cluster_steps

import (
	"context"

	"go.vervstack.ru/Velez/internal/pipelines/steps"
)

type preparePgDatabase struct {
}

func PreparePgDatabase() steps.Step {
	return &preparePgDatabase{}
}

func (p *preparePgDatabase) Do(ctx context.Context) error {
	return nil
}
