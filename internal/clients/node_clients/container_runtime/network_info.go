package container_runtime

import (
	"sort"
	"strings"

	"github.com/docker/docker/api/types/network"

	"go.vervstack.ru/Velez/internal/cluster/env"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

type networkState string

const (
	networkStateForeign          networkState = "foreign"
	networkStateManaged          networkState = "managed"
	networkStateOtherEnvironment networkState = "other_environment"
	networkStateSystem           networkState = "system"

	bridgeIccOption  = "com.docker.network.bridge.enable_icc"
	dockerShortIdLen = 12

	hostNetworkName = "host"
	noneNetworkName = "none"
	nullDriver      = "null"
)

func classifyNetwork(n network.Inspect, resolver nameResolver) networkState {
	switch n.Driver {
	case hostNetworkName, nullDriver:
		return networkStateSystem
	}

	switch n.Name {
	case defaultBridgeNetwork, hostNetworkName, noneNetworkName:
		return networkStateSystem
	}

	_, isLabelled := n.Labels[labels.NetworkManagedLabel]
	if isLabelled {
		if resolver.Owns(n.Labels) {
			return networkStateManaged
		}

		return networkStateOtherEnvironment
	}

	if n.Name == resolver.NetworkName(env.VervNetwork) {
		return networkStateManaged
	}

	return networkStateForeign
}

func buildNetworkInfo(n network.Inspect, state networkState, resolver nameResolver) NetworkInfo {
	isManaged := state == networkStateManaged

	name := n.Name
	if isManaged {
		name = resolver.VirtualContainerName(n.Name)
	}

	members := make([]NetworkMember, 0, len(n.Containers))
	for id, endpoint := range n.Containers {
		member := NetworkMember{
			ContainerId:   id,
			ContainerName: resolver.VirtualContainerName(strings.TrimPrefix(endpoint.Name, "/")),
			IpAddress:     stripIpMask(endpoint.IPv4Address),
		}

		members = append(members, member)
	}

	sort.Slice(members, func(i, j int) bool {
		return members[i].ContainerName < members[j].ContainerName
	})

	return NetworkInfo{
		Id:           n.ID,
		Name:         name,
		DockerName:   n.Name,
		Driver:       n.Driver,
		Subnet:       firstSubnet(n.IPAM),
		IsManaged:    isManaged,
		IsSystem:     state == networkStateSystem,
		IsInternal:   n.Internal,
		IsIccEnabled: n.Options[bridgeIccOption] != "false",
		Members:      members,
	}
}

func firstSubnet(ipam network.IPAM) string {
	for _, cfg := range ipam.Config {
		if cfg.Subnet != "" {
			return cfg.Subnet
		}
	}

	return ""
}

func stripIpMask(address string) string {
	ip, _, _ := strings.Cut(address, "/")

	return ip
}

// filterAliases drops the aliases Docker adds on its own: the container's short id.
func filterAliases(aliases []string, containerId, containerName string) []string {
	shortId := containerId
	if len(shortId) > dockerShortIdLen {
		shortId = shortId[:dockerShortIdLen]
	}

	name := strings.TrimPrefix(containerName, "/")

	filtered := make([]string, 0, len(aliases))
	for _, alias := range aliases {
		if alias == shortId || alias == name {
			continue
		}

		filtered = append(filtered, alias)
	}

	return filtered
}
