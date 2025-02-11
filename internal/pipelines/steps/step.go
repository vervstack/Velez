package steps

import (
	"context"
)

type Step interface {
	Do(ctx context.Context) error
}

type RollbackableStep interface {
	Rollback(ctx context.Context) error
}
