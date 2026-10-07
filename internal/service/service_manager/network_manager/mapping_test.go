package network_manager

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
)

func Test_NetworkCapabilities_Scenarios(t *testing.T) {
	tests := []struct {
		name     string
		info     container_runtime.NetworkInfo
		expected []velez_api.NetworkCapability
	}{
		{
			name: "managed network can be deleted and attached",
			info: container_runtime.NetworkInfo{IsManaged: true},
			expected: []velez_api.NetworkCapability{
				velez_api.NetworkCapability_NETWORK_CAPABILITY_DELETE,
				velez_api.NetworkCapability_NETWORK_CAPABILITY_ATTACH,
			},
		},
		{
			name:     "foreign network can only be attached",
			info:     container_runtime.NetworkInfo{},
			expected: []velez_api.NetworkCapability{velez_api.NetworkCapability_NETWORK_CAPABILITY_ATTACH},
		},
		{
			name:     "system network has no capabilities",
			info:     container_runtime.NetworkInfo{IsSystem: true},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, networkCapabilities(tt.info))
		})
	}
}

func Test_ToProtoNetwork_OptionalFieldsOmittedWhenEmpty(t *testing.T) {
	info := container_runtime.NetworkInfo{
		Id:      "net-1",
		Name:    "app",
		Members: []container_runtime.NetworkMember{{ContainerId: "c1", ContainerName: "web"}},
	}

	got := toProtoNetwork(info)

	require.Empty(t, got.GetSubnet())
	require.Len(t, got.GetMembers(), 1)
	require.Empty(t, got.GetMembers()[0].GetIpAddress())
	require.Equal(t, velez_api.NetworkProvider_NETWORK_PROVIDER_DOCKER, got.GetProvider())
}

func Test_ToProtoNetwork_OptionalFieldsSetWhenPresent(t *testing.T) {
	info := container_runtime.NetworkInfo{
		Id:           "net-1",
		Name:         "app",
		Subnet:       "172.20.0.0/16",
		IsManaged:    true,
		IsIccEnabled: true,
		Members: []container_runtime.NetworkMember{
			{ContainerId: "c1", ContainerName: "web", IpAddress: "172.20.0.2", Aliases: []string{"w"}},
		},
	}

	got := toProtoNetwork(info)

	require.Equal(t, "172.20.0.0/16", got.GetSubnet())
	require.Equal(t, "172.20.0.2", got.GetMembers()[0].GetIpAddress())
	require.Equal(t, []string{"w"}, got.GetMembers()[0].GetAliases())
	require.True(t, got.GetIsManaged())
	require.True(t, got.GetIsIccEnabled())
}

func Test_DockerProviderInfo_HasAllCapabilities(t *testing.T) {
	require.Len(t, dockerProviderInfo().GetCapabilities(), 4)
	require.Empty(t, vcnProviderInfo().GetCapabilities())
}

func Test_OptionalString_Scenarios(t *testing.T) {
	require.Nil(t, optionalString(""))

	got := optionalString("x")
	require.NotNil(t, got)
	require.Equal(t, "x", *got)
}
