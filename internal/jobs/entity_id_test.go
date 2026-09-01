package jobs_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/jobs"
)

// An empty suffix must leave the name untouched (the single-environment /
// unconfigured-ContainerSuffix node keeps its historical entity ids); a
// non-empty suffix must scope the id so two environments don't collide on
// velez.tasks' UNIQUE (entity_id, action).
func TestSmerdEntityID(t *testing.T) {
	t.Parallel()

	require.Equal(t, "svc", jobs.SmerdEntityID("", "svc"))
	require.Equal(t, "env/svc", jobs.SmerdEntityID("env", "svc"))
}
