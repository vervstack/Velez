package jobs

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service/user_errors"
)

const (
	testSwitchOldRootId = "oldrootid0123456789"
	testSwitchNewRootId = "newrootid0123456789"
	testSwitchNetwork   = "app_net"
	testSwitchPort      = nat.Port("80/tcp")
	testSwitchHostIp    = "0.0.0.0"
	testSwitchAlias     = "web"
)

func newSwitchPayload(rootName, rootId string) *velez_api.RegisterContainerTaskPayload {
	return &velez_api.RegisterContainerTaskPayload{
		ContainerId:               "serviceid",
		ContainerName:             testRolesService,
		RootContainerId:           rootId,
		RootContainerName:         rootName,
		PreviousRootContainerId:   testSwitchOldRootId,
		PreviousRootContainerName: testRolesRoot,
		SidecarContainerNames:     []string{testRolesRoot, testRolesOther},
	}
}

func Test_SwitchSidecarNames(t *testing.T) {
	t.Parallel()

	plan := networkGroupPlan{SidecarNames: []string{testRolesService, testRolesOther}}

	cases := []struct {
		name        string
		names       switchRoleNames
		isStatefull bool
		want        []string
	}{
		{
			name:  "single node, new root is the service",
			names: switchRoleNames{OldRoot: testRolesRoot, NewRoot: testRolesService, Service: testRolesService},
			want:  []string{testRolesRoot, testRolesOther},
		},
		{
			name:        "cluster, new root is the service",
			names:       switchRoleNames{OldRoot: testRolesRoot, NewRoot: testRolesService, Service: testRolesService},
			isStatefull: true,
			want:        []string{testRolesRoot, testRolesOther},
		},
		{
			name:  "single node, service is another member",
			names: switchRoleNames{OldRoot: testRolesRoot, NewRoot: testRolesService, Service: testRolesOther},
			want:  []string{testRolesRoot, testRolesOther},
		},
		{
			name:        "cluster, service is another member",
			names:       switchRoleNames{OldRoot: testRolesRoot, NewRoot: testRolesService, Service: testRolesOther},
			isStatefull: true,
			want:        []string{testRolesRoot, testRolesService},
		},
		{
			name:  "single node, old root is the service",
			names: switchRoleNames{OldRoot: testRolesRoot, NewRoot: testRolesService, Service: testRolesRoot},
			want:  []string{testRolesRoot, testRolesOther},
		},
		{
			name:        "cluster, old root is the service",
			names:       switchRoleNames{OldRoot: testRolesRoot, NewRoot: testRolesService, Service: testRolesRoot},
			isStatefull: true,
			want:        []string{testRolesOther, testRolesService},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := switchSidecarNames(plan, tc.names, tc.isStatefull)
			require.NoError(t, err)
			require.ElementsMatch(t, tc.want, got)
		})
	}
}

func Test_SwitchSidecarNames_RejectsUnadoptableRoles(t *testing.T) {
	t.Parallel()

	names := switchRoleNames{OldRoot: testRolesRoot, NewRoot: testRolesService, Service: testRolesService}

	t.Run("registered root", func(t *testing.T) {
		t.Parallel()

		plan := networkGroupPlan{IsRootRegistered: true, SidecarNames: []string{testRolesService}}

		_, err := switchSidecarNames(plan, names, false)
		require.ErrorIs(t, err, user_errors.ErrContainerAlreadyLinked)
	})

	t.Run("registered new root", func(t *testing.T) {
		t.Parallel()

		plan := networkGroupPlan{SidecarNames: []string{testRolesOther}}

		_, err := switchSidecarNames(plan, names, false)
		require.ErrorIs(t, err, user_errors.ErrContainerAlreadyLinked)
	})

	t.Run("registered service", func(t *testing.T) {
		t.Parallel()

		plan := networkGroupPlan{SidecarNames: []string{testRolesService}}
		other := switchRoleNames{OldRoot: testRolesRoot, NewRoot: testRolesService, Service: testRolesOther}

		_, err := switchSidecarNames(plan, other, false)
		require.ErrorIs(t, err, user_errors.ErrContainerAlreadyLinked)
	})
}

func Test_NewSwitchRootPlan_NewRootIsService(t *testing.T) {
	t.Parallel()

	payload := newSwitchPayload("", "")

	payload.SidecarContainerNames = []string{testRolesRoot, testRolesOther}

	plan := newSwitchRootPlan(payload)

	require.True(t, plan.IsNewRootService)
	require.Equal(t, "serviceid", plan.NewRootId)
	require.Equal(t, testRolesService, plan.NewRootName)
	require.Equal(t, testSwitchOldRootId, plan.OldRootId)
	require.Equal(t, []string{testRolesRoot, testRolesOther}, plan.Members)
}

func Test_NewSwitchRootPlan_NewRootIsNotService(t *testing.T) {
	t.Parallel()

	payload := newSwitchPayload(testRolesOther, testSwitchNewRootId)

	payload.SidecarContainerNames = []string{testRolesRoot}

	plan := newSwitchRootPlan(payload)

	require.False(t, plan.IsNewRootService)
	require.Equal(t, testSwitchNewRootId, plan.NewRootId)
	require.Equal(t, testRolesOther, plan.NewRootName)
	require.Equal(t, []string{testRolesRoot, testRolesService}, plan.Members)
}

func Test_NewSwitchRootPlan_OldRootAsServiceIsListedOnce(t *testing.T) {
	t.Parallel()

	payload := newSwitchPayload(testRolesOther, testSwitchNewRootId)

	payload.ContainerName = testRolesRoot

	plan := newSwitchRootPlan(payload)

	require.Equal(t, []string{testRolesRoot}, plan.Members)
}

func newSwitchOldRoot() container.InspectResponse {
	return container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{
			ID:   testSwitchOldRootId,
			Name: "/" + testRolesRoot,
			HostConfig: &container.HostConfig{
				NetworkMode:  testSwitchNetwork,
				PortBindings: nat.PortMap{testSwitchPort: {{HostPort: ""}}},
			},
		},
		Config:          &container.Config{Image: "nginx:1", ExposedPorts: nat.PortSet{testSwitchPort: {}}},
		NetworkSettings: newSwitchNetworkSettings(),
	}
}

func newSwitchNetworkSettings() *container.NetworkSettings {
	settings := &container.NetworkSettings{
		Networks: map[string]*network.EndpointSettings{
			testSwitchNetwork: {
				NetworkID: "net1",
				Aliases:   []string{testSwitchAlias, testSwitchOldRootId[:12], testSwitchAlias},
			},
			"extra_net": {NetworkID: "net2", Aliases: []string{"extra"}},
		},
	}

	settings.Ports = nat.PortMap{testSwitchPort: {{HostIP: testSwitchHostIp, HostPort: "32768"}}}

	return settings
}

func newSwitchNewRoot() container.InspectResponse {
	return container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{
			ID:   testSwitchNewRootId,
			Name: "/" + testRolesService,
			HostConfig: &container.HostConfig{
				NetworkMode: container.NetworkMode(networkModeOfPrefix + testSwitchOldRootId),
				DNS:         []string{"1.1.1.1"},
			},
		},
		Config: &container.Config{
			Image:    "tailscale:1",
			Hostname: "host",
			Labels:   map[string]string{testKeepLabel: "me", labels.DisplayNameLabel: testStaleOld},
		},
	}
}

func Test_MovedPortBindings_UsesPublishedHostPorts(t *testing.T) {
	t.Parallel()

	bindings := movedPortBindings(newSwitchOldRoot())

	require.Equal(t, nat.PortMap{testSwitchPort: {{HostIP: testSwitchHostIp, HostPort: "32768"}}}, bindings)
}

func Test_MovedPortBindings_FallsBackToConfiguredBindings(t *testing.T) {
	t.Parallel()

	oldRoot := newSwitchOldRoot()

	oldRoot.NetworkSettings = nil

	require.Equal(t, oldRoot.HostConfig.PortBindings, movedPortBindings(oldRoot))
}

func Test_MovedAliases_DropsShortIdAndKeepsOldName(t *testing.T) {
	t.Parallel()

	oldRoot := newSwitchOldRoot()

	aliases := movedAliases(oldRoot, oldRoot.NetworkSettings.Networks[testSwitchNetwork])

	require.Equal(t, []string{testSwitchAlias, testRolesRoot}, aliases)
}

func Test_MovedNetworkingConfig(t *testing.T) {
	t.Parallel()

	t.Run("user defined network carries aliases", func(t *testing.T) {
		t.Parallel()

		config := movedNetworkingConfig(newSwitchOldRoot())
		require.NotNil(t, config)
		require.Equal(t, []string{testSwitchAlias, testRolesRoot}, config.EndpointsConfig[testSwitchNetwork].Aliases)
		require.NotContains(t, config.EndpointsConfig, "extra_net")
	})

	t.Run("default bridge gets no aliases", func(t *testing.T) {
		t.Parallel()

		oldRoot := newSwitchOldRoot()

		oldRoot.HostConfig.NetworkMode = defaultNetworkMode
		oldRoot.NetworkSettings.Networks[defaultBridgeNetwork] = &network.EndpointSettings{NetworkID: "bridge1"}

		config := movedNetworkingConfig(oldRoot)
		require.NotNil(t, config)
		require.Empty(t, config.EndpointsConfig[defaultBridgeNetwork].Aliases)
	})

	t.Run("host network has no endpoint", func(t *testing.T) {
		t.Parallel()

		oldRoot := newSwitchOldRoot()

		oldRoot.HostConfig.NetworkMode = hostNetwork

		require.Nil(t, movedNetworkingConfig(oldRoot))
	})
}

func Test_NewRootCreateRequest_TakesOverOldRootNetwork(t *testing.T) {
	t.Parallel()

	oldRoot := newSwitchOldRoot()
	plan := switchRootPlan{OldRootId: testSwitchOldRootId, NewRootName: testRolesService, IsNewRootService: true}
	stamped := switchedRootLabels(testSwitchAlias, plan, testSwitchOldRootId)

	req := newRootCreateRequest(testRolesService, stamped, oldRoot, newSwitchNewRoot())

	require.Equal(t, testRolesService, req.ContainerName)
	require.Equal(t, container.NetworkMode(testSwitchNetwork), req.HostConfig.NetworkMode)
	require.Equal(t, oldRoot.NetworkSettings.Ports, req.HostConfig.PortBindings)
	require.Equal(t, oldRoot.Config.ExposedPorts, req.Config.ExposedPorts)
	require.Empty(t, req.HostConfig.DNS, "the new root's own DNS is replaced by the old root's")
	require.Equal(t, "tailscale:1", req.Config.Image)
	require.Equal(t, testSwitchAlias, req.Config.Labels[labels.VervServiceLabel])
	require.Equal(t, testSwitchAlias, req.Config.Labels[labels.DisplayNameLabel])
	require.Equal(t, testSwitchOldRootId, req.Config.Labels[labels.OnboardedFromLabel])
	require.Equal(t, "me", req.Config.Labels[testKeepLabel])
	require.NotNil(t, req.NetworkingConfig)
}

func Test_SwitchedRootLabels(t *testing.T) {
	t.Parallel()

	t.Run("service root has display name and no sidecar marker", func(t *testing.T) {
		t.Parallel()

		stamped := switchedRootLabels(testSwitchAlias, switchRootPlan{IsNewRootService: true}, testSwitchOldRootId)

		require.Contains(t, stamped, labels.DisplayNameLabel)
		require.NotContains(t, stamped, labels.Sidecar)
		require.Equal(t, testSwitchOldRootId, stamped[labels.OnboardedFromLabel])
	})

	t.Run("resource root is a sidecar without display name", func(t *testing.T) {
		t.Parallel()

		stamped := switchedRootLabels(testSwitchAlias, switchRootPlan{}, testSwitchOldRootId)

		require.NotContains(t, stamped, labels.DisplayNameLabel)
		require.Contains(t, stamped, labels.Sidecar)
		require.Equal(t, testSwitchOldRootId, stamped[labels.OnboardedFromLabel])
	})
}
