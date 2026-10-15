package transfer_in

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
)

const (
	testNetwork   = "app_net"
	testContainer = "web"
)

func newNetworkedContainer() manifestContainer {
	inspect := container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{
			ID: "0123456789abcdef",
			HostConfig: &container.HostConfig{
				NetworkMode:  testNetwork,
				PortBindings: nat.PortMap{"80/tcp": {{HostPort: "8080"}}},
			},
		},
		Config: &container.Config{Hostname: "0123456789ab", Image: "nginx"},
		NetworkSettings: &container.NetworkSettings{Networks: map[string]*network.EndpointSettings{
			testNetwork: {
				Aliases:    []string{testContainer, "0123456789ab"},
				IPAddress:  "172.18.0.2",
				MacAddress: "02:42:ac:12:00:02",
				NetworkID:  "net-id",
				EndpointID: "endpoint-id",
			},
			networkBridge: {Aliases: []string{"ignored"}},
		}},
	}

	return manifestContainer{name: testContainer, inspect: inspect}
}

func Test_RootCreateRequest_KeepsOnlyAliasesOfUserNetworks(t *testing.T) {
	req := rootCreateRequest(newNetworkedContainer())

	require.Equal(t, testContainer, req.ContainerName)
	require.Empty(t, req.Config.Hostname)
	require.Equal(t, "nginx", req.Config.Image)
	require.Len(t, req.NetworkingConfig.EndpointsConfig, 1)

	endpoint := req.NetworkingConfig.EndpointsConfig[testNetwork]
	require.Equal(t, &network.EndpointSettings{Aliases: []string{testContainer}}, endpoint)
	require.NotEmpty(t, req.HostConfig.PortBindings)
}

func Test_SidecarCreateRequest_JoinsRootNamespace(t *testing.T) {
	req := sidecarCreateRequest(newNetworkedContainer(), "root-id")

	require.Equal(t, container.NetworkMode("container:root-id"), req.HostConfig.NetworkMode)
	require.Empty(t, req.HostConfig.PortBindings)
	require.Empty(t, req.Config.Hostname)
	require.Nil(t, req.NetworkingConfig)
}

func Test_NetworkNames_Scenarios(t *testing.T) {
	names := networkNames(newNetworkedContainer().inspect)

	require.Equal(t, []string{testNetwork}, names)
}

func Test_IsUserNetwork_Scenarios(t *testing.T) {
	cases := []struct {
		name string
		want bool
	}{
		{testNetwork, true},
		{networkBridge, false},
		{networkHost, false},
		{networkNone, false},
		{"container:abc", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, isUserNetwork(tc.name))
		})
	}
}
