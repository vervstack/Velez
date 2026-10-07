package jobs

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	testReconcileEnvironment = "default"
	testStateExited          = "exited"
)

func newTestSummary(id, state string, containerLabels map[string]string) container.Summary {
	return container.Summary{ID: id, State: state, Labels: containerLabels}
}

func serviceLabels(name string) map[string]string {
	return map[string]string{labels.VervServiceLabel: name}
}

func Test_RunningPrimaries_Grouping(t *testing.T) {
	sidecarLabels := serviceLabels("api")

	sidecarLabels[labels.Sidecar] = vervConfigLabelEnabled

	cases := []struct {
		name      string
		summaries []container.Summary
		wantIds   []string
	}{
		{
			name:      "running service is a primary",
			summaries: []container.Summary{newTestSummary("a", containerStateRunning, serviceLabels("api"))},
			wantIds:   []string{"a"},
		},
		{
			name:      "sidecar is skipped",
			summaries: []container.Summary{newTestSummary("a", containerStateRunning, sidecarLabels)},
		},
		{
			name:      "unlabelled container is skipped",
			summaries: []container.Summary{newTestSummary("a", containerStateRunning, map[string]string{})},
		},
		{
			name: "running container is preferred over a stopped one",
			summaries: []container.Summary{
				newTestSummary("old", testStateExited, serviceLabels("api")),
				newTestSummary("new", containerStateRunning, serviceLabels("api")),
			},
			wantIds: []string{"new"},
		},
		{
			name:      "stopped-only service is excluded",
			summaries: []container.Summary{newTestSummary("a", testStateExited, serviceLabels("api"))},
		},
		{
			name: "services keep first-seen order",
			summaries: []container.Summary{
				newTestSummary("b", containerStateRunning, serviceLabels("b")),
				newTestSummary("a", containerStateRunning, serviceLabels("a")),
			},
			wantIds: []string{"b", "a"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			listed := environmentSummaries{environment: testReconcileEnvironment, summaries: tc.summaries}

			got := runningPrimaries(listed)

			var gotIds []string

			for _, primary := range got {
				gotIds = append(gotIds, primary.summary.ID)
				require.Equal(t, testReconcileEnvironment, primary.environment)
			}

			require.Equal(t, tc.wantIds, gotIds)
		})
	}
}
