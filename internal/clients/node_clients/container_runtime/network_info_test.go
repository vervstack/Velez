package container_runtime

import (
	"testing"

	"github.com/docker/docker/api/types/network"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	testNetworkEnvironmentSuffix = "stage"
	testBridgeDriver             = "bridge"
)

func Test_ClassifyNetwork_Scenarios(t *testing.T) {
	resolver := &labelSuffixResolver{suffix: testNetworkEnvironmentSuffix}

	managedLabels := func(suffix string) map[string]string {
		return map[string]string{
			labels.NetworkManagedLabel: labels.NetworkManagedLabelValue,
			labels.SuffixLabel:         suffix,
		}
	}

	tests := []struct {
		name     string
		network  network.Inspect
		expected networkState
	}{
		{
			name:     "default bridge is system",
			network:  network.Inspect{Name: defaultBridgeNetwork, Driver: testBridgeDriver},
			expected: networkStateSystem,
		},
		{
			name:     "host driver is system",
			network:  network.Inspect{Name: "anything", Driver: hostNetworkName},
			expected: networkStateSystem,
		},
		{
			name:     "null driver is system",
			network:  network.Inspect{Name: "anything", Driver: "null"},
			expected: networkStateSystem,
		},
		{
			name:     "none name is system",
			network:  network.Inspect{Name: noneNetworkName, Driver: testBridgeDriver},
			expected: networkStateSystem,
		},
		{
			name:     "labelled network of this environment is managed",
			network:  network.Inspect{Name: "app_stage", Driver: testBridgeDriver, Labels: managedLabels("stage")},
			expected: networkStateManaged,
		},
		{
			name:     "labelled network of another environment is skipped",
			network:  network.Inspect{Name: "app_prod", Driver: testBridgeDriver, Labels: managedLabels("prod")},
			expected: networkStateOtherEnvironment,
		},
		{
			name:     "unlabelled legacy verv network of this environment is managed",
			network:  network.Inspect{Name: "verv_stage", Driver: testBridgeDriver},
			expected: networkStateManaged,
		},
		{
			name:     "unlabelled verv network without suffix is foreign here",
			network:  network.Inspect{Name: "verv", Driver: testBridgeDriver},
			expected: networkStateForeign,
		},
		{
			name:     "unlabelled arbitrary network is foreign",
			network:  network.Inspect{Name: "compose_default", Driver: testBridgeDriver},
			expected: networkStateForeign,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, classifyNetwork(tt.network, resolver))
		})
	}
}

func Test_ClassifyNetwork_DirectResolverOwnsEveryLabelledNetwork(t *testing.T) {
	inspected := network.Inspect{
		Name:   "app",
		Driver: testBridgeDriver,
		Labels: map[string]string{labels.NetworkManagedLabel: labels.NetworkManagedLabelValue},
	}

	require.Equal(t, networkStateManaged, classifyNetwork(inspected, &directResolver{}))
}

func Test_BuildNetworkInfo_ManagedNetwork(t *testing.T) {
	resolver := &labelSuffixResolver{suffix: testNetworkEnvironmentSuffix}

	inspected := network.Inspect{
		ID:       "net-1",
		Name:     "app_stage",
		Driver:   "bridge",
		Internal: true,
		Options:  map[string]string{bridgeIccOption: "false"},
		IPAM: network.IPAM{
			Config: []network.IPAMConfig{{Subnet: ""}, {Subnet: "172.20.0.0/16"}},
		},
		Containers: map[string]network.EndpointResource{
			"c2": {Name: "zeta_stage", IPv4Address: "172.20.0.3/16"},
			"c1": {Name: "/alpha_stage", IPv4Address: "172.20.0.2/16"},
			"c3": {Name: "noip"},
		},
	}

	info := buildNetworkInfo(inspected, networkStateManaged, resolver)

	require.Equal(t, "net-1", info.Id)
	require.Equal(t, "app", info.Name)
	require.Equal(t, "app_stage", info.DockerName)
	require.Equal(t, "bridge", info.Driver)
	require.Equal(t, "172.20.0.0/16", info.Subnet)
	require.True(t, info.IsManaged)
	require.False(t, info.IsSystem)
	require.True(t, info.IsInternal)
	require.False(t, info.IsIccEnabled)

	require.Len(t, info.Members, 3)
	require.Equal(t, NetworkMember{ContainerId: "c1", ContainerName: "alpha", IpAddress: "172.20.0.2"}, info.Members[0])
	require.Equal(t, "noip", info.Members[1].ContainerName)
	require.Empty(t, info.Members[1].IpAddress)
	require.Equal(t, "zeta", info.Members[2].ContainerName)
}

func Test_BuildNetworkInfo_ForeignNetworkKeepsDockerName(t *testing.T) {
	resolver := &labelSuffixResolver{suffix: testNetworkEnvironmentSuffix}

	inspected := network.Inspect{ID: "net-2", Name: "compose_default", Driver: testBridgeDriver}

	info := buildNetworkInfo(inspected, networkStateForeign, resolver)

	require.Equal(t, "compose_default", info.Name)
	require.False(t, info.IsManaged)
	require.True(t, info.IsIccEnabled)
	require.Empty(t, info.Subnet)
	require.Empty(t, info.Members)
}

func Test_StripIpMask_Scenarios(t *testing.T) {
	require.Equal(t, "10.0.0.2", stripIpMask("10.0.0.2/24"))
	require.Equal(t, "10.0.0.2", stripIpMask("10.0.0.2"))
	require.Empty(t, stripIpMask(""))
}

func Test_FilterAliases_DropsDockerGeneratedOnes(t *testing.T) {
	containerId := "0123456789abcdef0123456789abcdef"

	aliases := []string{"0123456789ab", "web", "my_container"}

	got := filterAliases(aliases, containerId, "/my_container")

	require.Equal(t, []string{"web"}, got)
}
