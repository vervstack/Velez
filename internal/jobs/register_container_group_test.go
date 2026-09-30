package jobs

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	testGroupService = "vpn-app"
	testKeepLabel    = "keep"
	testKeepValue    = "me"
)

func newSidecarSummary(name string, containerLabels map[string]string) container.Summary {
	return container.Summary{
		Names:  []string{"/" + name},
		Labels: containerLabels,
	}
}

func Test_PlanNetworkGroup_UnregisteredRootAdoptsEverySidecar(t *testing.T) {
	t.Parallel()

	sidecars := []container.Summary{
		newSidecarSummary("ts-a", nil),
		newSidecarSummary("ts-b", map[string]string{testKeepLabel: testKeepValue}),
	}

	plan := planNetworkGroup(map[string]string{}, "", sidecars, map[string]string{})

	require.False(t, plan.IsRootRegistered)
	require.Empty(t, plan.ServiceName)
	require.Equal(t, []string{"ts-a", "ts-b"}, plan.SidecarNames)
}

func Test_PlanNetworkGroup_NoSidecarsIsEmptyPlan(t *testing.T) {
	t.Parallel()

	plan := planNetworkGroup(map[string]string{}, "", nil, map[string]string{})

	require.False(t, plan.IsRootRegistered)
	require.Empty(t, plan.SidecarNames)
}

func Test_PlanNetworkGroup_LabelledRootKeepsItsService(t *testing.T) {
	t.Parallel()

	rootLabels := map[string]string{labels.VervServiceLabel: testGroupService}
	sidecars := []container.Summary{newSidecarSummary("ts-a", nil)}

	plan := planNetworkGroup(rootLabels, "other", sidecars, map[string]string{})

	require.True(t, plan.IsRootRegistered)
	require.Equal(t, testGroupService, plan.ServiceName)
	require.Equal(t, []string{"ts-a"}, plan.SidecarNames)
}

func Test_PlanNetworkGroup_BoundRootKeepsItsBindingService(t *testing.T) {
	t.Parallel()

	sidecars := []container.Summary{newSidecarSummary("ts-a", nil)}

	plan := planNetworkGroup(map[string]string{}, testGroupService, sidecars, map[string]string{})

	require.True(t, plan.IsRootRegistered)
	require.Equal(t, testGroupService, plan.ServiceName)
}

func Test_PlanNetworkGroup_RegisteredSidecarsAreSkipped(t *testing.T) {
	t.Parallel()

	registeredLabels := map[string]string{
		labels.Sidecar:          labelTrueValue,
		labels.VervServiceLabel: testGroupService,
	}

	sidecars := []container.Summary{
		newSidecarSummary("by-labels", registeredLabels),
		newSidecarSummary("by-binding", nil),
		newSidecarSummary("pending", nil),
	}
	bound := map[string]string{"by-binding": testGroupService}

	plan := planNetworkGroup(map[string]string{}, "", sidecars, bound)

	require.Equal(t, []string{"pending"}, plan.SidecarNames)
}

func Test_IsSidecarRegistered_Cases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		labels        map[string]string
		boundService  string
		wantRegistred bool
	}{
		{"no labels no binding", nil, "", false},
		{"sidecar marker without service", map[string]string{labels.Sidecar: labelTrueValue}, "", false},
		{"service without sidecar marker", map[string]string{labels.VervServiceLabel: testSvc}, "", false},
		{"marker and service", map[string]string{labels.Sidecar: labelTrueValue, labels.VervServiceLabel: testSvc}, "", true},
		{"binding only", nil, testSvc, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.wantRegistred, isSidecarRegistered(tc.labels, tc.boundService))
		})
	}
}

type fakeRootRegistered struct {
	isRegistered bool
}

func (f fakeRootRegistered) GetIsRootRegistered() bool {
	return f.isRegistered
}

type countingJob struct {
	doCalls       int
	rollbackCalls int
}

func (j *countingJob) Do(context.Context) error {
	j.doCalls++

	return nil
}

func (j *countingJob) Rollback(context.Context) error {
	j.rollbackCalls++

	return nil
}

func Test_UnlessRootRegistered_SkipsInnerWhenRootRegistered(t *testing.T) {
	t.Parallel()

	inner := &countingJob{}
	job := &unlessRootRegistered{inner: inner, group: fakeRootRegistered{isRegistered: true}}

	require.NoError(t, job.Do(context.Background()))
	require.NoError(t, job.Rollback(context.Background()))
	require.Zero(t, inner.doCalls)
	require.Zero(t, inner.rollbackCalls)
}

func Test_UnlessRootRegistered_RunsInnerWhenRootUnregistered(t *testing.T) {
	t.Parallel()

	inner := &countingJob{}
	job := &unlessRootRegistered{inner: inner, group: fakeRootRegistered{}}

	require.NoError(t, job.Do(context.Background()))
	require.NoError(t, job.Rollback(context.Background()))
	require.Equal(t, 1, inner.doCalls)
	require.Equal(t, 1, inner.rollbackCalls)
}
