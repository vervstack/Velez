package jobs

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain/labels"
)

func Test_SidecarLabels_CarryServiceAndSidecarMarker(t *testing.T) {
	t.Parallel()

	got := sidecarLabels(testGroupService)

	require.Equal(t, map[string]string{
		labels.VervServiceLabel: testGroupService,
		labels.DisplayNameLabel: testGroupService,
		labels.Sidecar:          labelTrueValue,
	}, got)
}

func Test_IsRegisteredSidecar_MatchesOnlyOwnService(t *testing.T) {
	t.Parallel()

	stamped := sidecarLabels(testGroupService)

	require.True(t, isRegisteredSidecar(stamped, testGroupService))
	require.False(t, isRegisteredSidecar(stamped, "other"))
	require.False(t, isRegisteredSidecar(map[string]string{labels.VervServiceLabel: testGroupService}, testGroupService))
}

func newInspectedSidecar() container.InspectResponse {
	config := &container.Config{
		Image:        "tailscale/tailscale:latest",
		Hostname:     "abc123",
		Domainname:   "example.org",
		Env:          []string{"TS_AUTHKEY=secret"},
		ExposedPorts: nat.PortSet{"41641/udp": struct{}{}},
		Labels:       map[string]string{testKeepLabel: testKeepValue, labels.DisplayNameLabel: testStaleOld},
	}

	hostConfig := &container.HostConfig{
		NetworkMode:     "container:oldroot",
		Binds:           []string{"ts-state:/var/lib/tailscale"},
		PortBindings:    nat.PortMap{"41641/udp": {{HostPort: "41641"}}},
		PublishAllPorts: true,
		RestartPolicy:   container.RestartPolicy{Name: container.RestartPolicyAlways},
	}

	return container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{HostConfig: hostConfig},
		Config:            config,
	}
}

func Test_SidecarCreateRequest_JoinsNewRootAndKeepsSettings(t *testing.T) {
	t.Parallel()

	source := newInspectedSidecar()

	req := sidecarCreateRequest("ts", testGroupService, "container:newroot", source)

	require.Equal(t, "ts", req.ContainerName)
	require.Equal(t, container.NetworkMode("container:newroot"), req.HostConfig.NetworkMode)
	require.Nil(t, req.NetworkingConfig)
	require.Empty(t, req.HostConfig.PortBindings)
	require.False(t, req.HostConfig.PublishAllPorts)
	require.Equal(t, []string{"ts-state:/var/lib/tailscale"}, req.HostConfig.Binds)
	require.Equal(t, container.RestartPolicyAlways, req.HostConfig.RestartPolicy.Name)
	require.Equal(t, "tailscale/tailscale:latest", req.Config.Image)
	require.Equal(t, []string{"TS_AUTHKEY=secret"}, req.Config.Env)
	require.Empty(t, req.Config.Hostname)
	require.Empty(t, req.Config.Domainname)
	require.Empty(t, req.Config.ExposedPorts)
}

func Test_SidecarCreateRequest_MergesLabelsAndLeavesSourceUntouched(t *testing.T) {
	t.Parallel()

	source := newInspectedSidecar()

	req := sidecarCreateRequest("ts", testGroupService, "container:newroot", source)

	require.Equal(t, map[string]string{
		testKeepLabel:           testKeepValue,
		labels.VervServiceLabel: testGroupService,
		labels.DisplayNameLabel: testGroupService,
		labels.Sidecar:          labelTrueValue,
	}, req.Config.Labels)

	require.Equal(t, "old", source.Config.Labels[labels.DisplayNameLabel])
	require.Equal(t, container.NetworkMode("container:oldroot"), source.HostConfig.NetworkMode)
	require.Equal(t, "abc123", source.Config.Hostname)
}
