package jobs

import (
	"context"
)

type internalEnqueueKey struct{}

// WithInternalEnqueue marks ctx as belonging to work that must keep enqueuing
// while the engine drains: a running task enqueuing its children, or the
// deploy watcher feeding a task that waits on them.
func WithInternalEnqueue(ctx context.Context) context.Context {
	return context.WithValue(ctx, internalEnqueueKey{}, true)
}

func isInternalEnqueue(ctx context.Context) bool {
	isInternal, ok := ctx.Value(internalEnqueueKey{}).(bool)

	return ok && isInternal
}
