package jobs

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service/user_errors"
)

const (
	testRolesRoot    = "app"
	testRolesService = "ts"
	testRolesOther   = "metrics"
)

func Test_RegisterSidecarNames_ServiceIsRootKeepsPlan(t *testing.T) {
	t.Parallel()

	plan := networkGroupPlan{SidecarNames: []string{testRolesService, testRolesOther}}

	for _, serviceName := range []string{"", testRolesRoot} {
		for _, isStatefull := range []bool{false, true} {
			names, err := registerSidecarNames(plan, testRolesRoot, serviceName, isStatefull)
			require.NoError(t, err)
			require.Equal(t, plan.SidecarNames, names)
		}
	}
}

func Test_RegisterSidecarNames_ServiceOtherThanRoot(t *testing.T) {
	t.Parallel()

	plan := networkGroupPlan{SidecarNames: []string{testRolesOther, testRolesService}}

	cases := []struct {
		name        string
		isStatefull bool
		want        []string
	}{
		{"single node recreates every member including the service", false, []string{testRolesOther, testRolesService}},
		{"cluster binds the root and the rest but not the service", true, []string{testRolesRoot, testRolesOther}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			names, err := registerSidecarNames(plan, testRolesRoot, testRolesService, tc.isStatefull)
			require.NoError(t, err)
			require.Equal(t, tc.want, names)
		})
	}
}

func Test_RegisterSidecarNames_RejectsUnadoptableService(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		plan networkGroupPlan
	}{
		{"service already registered", networkGroupPlan{SidecarNames: []string{testRolesOther}}},
		{"root already registered", networkGroupPlan{IsRootRegistered: true, SidecarNames: []string{testRolesService}}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := registerSidecarNames(tc.plan, testRolesRoot, testRolesService, false)
			require.ErrorIs(t, err, user_errors.ErrContainerAlreadyLinked)
		})
	}
}

func Test_RegisterSidecarGroup_RootIsServiceUsesContainerName(t *testing.T) {
	t.Parallel()

	group := registerSidecarGroup{payload: &velez_api.RegisterContainerTaskPayload{ContainerName: testRolesRoot}}

	require.Equal(t, testRolesRoot, group.GetContainerName())
	require.Empty(t, group.GetServiceMemberName())
}

func Test_RegisterSidecarGroup_ServiceOtherThanRootNamesRootAndMember(t *testing.T) {
	t.Parallel()

	group := registerSidecarGroup{payload: &velez_api.RegisterContainerTaskPayload{
		ContainerName:     testRolesService,
		RootContainerName: testRolesRoot,
	}}

	require.Equal(t, testRolesRoot, group.GetContainerName())
	require.Equal(t, testRolesService, group.GetServiceMemberName())
}

func Test_MemberLabelsFor_StampsServiceMemberWithoutSidecarMarker(t *testing.T) {
	t.Parallel()

	group := registerSidecarGroup{payload: &velez_api.RegisterContainerTaskPayload{
		ServiceName:       testGroupService,
		ContainerName:     testRolesService,
		RootContainerName: testRolesRoot,
	}}

	member := memberLabelsFor(group, testRolesService)
	require.Equal(t, map[string]string{
		labels.VervServiceLabel: testGroupService,
		labels.DisplayNameLabel: testGroupService,
	}, member)

	plain := memberLabelsFor(group, testRolesOther)
	require.Equal(t, sidecarLabels(testGroupService), plain)
}

func Test_MemberLabelsFor_RootIsServiceStampsEverySidecarPlain(t *testing.T) {
	t.Parallel()

	group := registerSidecarGroup{payload: &velez_api.RegisterContainerTaskPayload{
		ServiceName:   testGroupService,
		ContainerName: testRolesRoot,
	}}

	require.Equal(t, sidecarLabels(testGroupService), memberLabelsFor(group, testRolesService))
}

func Test_IsRegisteredMemberOf_RecognisesWhatMemberLabelsForStamps(t *testing.T) {
	t.Parallel()

	group := registerSidecarGroup{payload: &velez_api.RegisterContainerTaskPayload{
		ServiceName:       testGroupService,
		ContainerName:     testRolesService,
		RootContainerName: testRolesRoot,
	}}

	serviceStamp := memberLabelsFor(group, testRolesService)
	sidecarStamp := memberLabelsFor(group, testRolesOther)

	require.True(t, isRegisteredMemberOf(group, testRolesService, serviceStamp))
	require.False(t, isRegisteredMemberOf(group, testRolesService, sidecarStamp))
	require.True(t, isRegisteredMemberOf(group, testRolesOther, sidecarStamp))
	require.False(t, isRegisteredMemberOf(group, testRolesOther, serviceStamp))
}

func Test_MemberCreateRequest_ServiceMemberKeepsStampedDisplayName(t *testing.T) {
	t.Parallel()

	source := newInspectedSidecar()

	req := memberCreateRequest("ts", serviceMemberLabels(testGroupService), "container:newroot", source)

	require.Equal(t, testGroupService, req.Config.Labels[labels.DisplayNameLabel])
	require.Equal(t, testGroupService, req.Config.Labels[labels.VervServiceLabel])
	require.NotContains(t, req.Config.Labels, labels.Sidecar)
	require.Equal(t, testKeepValue, req.Config.Labels[testKeepLabel])
}

func Test_PlanNetworkGroup_RecreatedServiceMemberIsRegistered(t *testing.T) {
	t.Parallel()

	stamped := serviceMemberLabels(testGroupService)
	sidecars := []container.Summary{newSidecarSummary(testRolesService, stamped)}

	plan := planNetworkGroup(map[string]string{}, "", sidecars, map[string]string{})

	require.Empty(t, plan.SidecarNames)
}
