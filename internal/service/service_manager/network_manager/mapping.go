package network_manager

import (
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
)

func toProtoNetwork(info container_runtime.NetworkInfo) *velez_api.Network {
	members := make([]*velez_api.NetworkMember, 0, len(info.Members))
	for _, member := range info.Members {
		members = append(members, toProtoMember(member))
	}

	return &velez_api.Network{
		Id:           info.Id,
		Name:         info.Name,
		Provider:     velez_api.NetworkProvider_NETWORK_PROVIDER_DOCKER,
		IsManaged:    info.IsManaged,
		IsInternal:   info.IsInternal,
		IsIccEnabled: info.IsIccEnabled,
		Subnet:       optionalString(info.Subnet),
		Members:      members,
		Capabilities: networkCapabilities(info),
	}
}

func toProtoMember(member container_runtime.NetworkMember) *velez_api.NetworkMember {
	return &velez_api.NetworkMember{
		ContainerId:   member.ContainerId,
		ContainerName: member.ContainerName,
		Aliases:       member.Aliases,
		IpAddress:     optionalString(member.IpAddress),
	}
}

func networkCapabilities(info container_runtime.NetworkInfo) []velez_api.NetworkCapability {
	switch {
	case info.IsManaged:
		return []velez_api.NetworkCapability{
			velez_api.NetworkCapability_NETWORK_CAPABILITY_DELETE,
			velez_api.NetworkCapability_NETWORK_CAPABILITY_ATTACH,
		}
	case info.IsSystem:
		return nil
	default:
		return []velez_api.NetworkCapability{velez_api.NetworkCapability_NETWORK_CAPABILITY_ATTACH}
	}
}

func dockerProviderInfo() *velez_api.NetworkProviderInfo {
	return &velez_api.NetworkProviderInfo{
		Provider: velez_api.NetworkProvider_NETWORK_PROVIDER_DOCKER,
		Capabilities: []velez_api.NetworkCapability{
			velez_api.NetworkCapability_NETWORK_CAPABILITY_CREATE,
			velez_api.NetworkCapability_NETWORK_CAPABILITY_DELETE,
			velez_api.NetworkCapability_NETWORK_CAPABILITY_ATTACH,
			velez_api.NetworkCapability_NETWORK_CAPABILITY_RESTRICT,
		},
	}
}

func vcnProviderInfo() *velez_api.NetworkProviderInfo {
	return &velez_api.NetworkProviderInfo{
		Provider: velez_api.NetworkProvider_NETWORK_PROVIDER_VCN,
	}
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}
