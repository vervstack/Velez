package configuration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/cluster/env/container_service_task"
)

func Test_SharedInstanceContext_RoundTrip(t *testing.T) {
	ctx := context.Background()

	_, ok := sharedTaskFromContext(ctx)
	require.False(t, ok)

	task := &container_service_task.TaskV2{}
	instance := &SharedInstance{task: task}

	ctx = WithSharedInstance(ctx, instance)

	got, ok := sharedTaskFromContext(ctx)
	require.True(t, ok)
	require.Same(t, task, got)
}
