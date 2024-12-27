package shared

import (
	"context"
)

type Step interface {
	Do(ctx context.Context) error
}

type RollbackableStep interface {
	Rollback(_ context.Context) error
}
