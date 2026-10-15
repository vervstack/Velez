package transfer_in

import (
	"slices"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"

	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
)

const (
	shortIdLen           = 12
	containerNetworkMode = "container:"

	networkBridge = "bridge"
	networkHost   = "host"
	networkNone   = "none"
)

// rootCreateRequest rebuilds a root container on this node: it keeps the
// network names and aliases but none of the addresses the source node assigned.
func rootCreateRequest(c manifestContainer) container_runtime.ContainerCreateRequest {
	config := *c.inspect.Config

	if config.Hostname == shortId(c.inspect.ID) {
		config.Hostname = ""
	}

	hostConfig := *c.inspect.HostConfig

	networking := &network.NetworkingConfig{EndpointsConfig: endpointsConfig(c.inspect)}

	return container_runtime.ContainerCreateRequest{
		Config:           &container_runtime.ContainerConfig{Config: &config},
		HostConfig:       &container_runtime.HostConfig{HostConfig: &hostConfig},
		NetworkingConfig: &container_runtime.NetworkingConfig{NetworkingConfig: networking},
		ContainerName:    c.name,
	}
}

// sidecarCreateRequest joins the sidecar to its root's network namespace.
// Settings Docker refuses next to a container network mode are dropped.
func sidecarCreateRequest(c manifestContainer, rootId string) container_runtime.ContainerCreateRequest {
	config := *c.inspect.Config

	config.Hostname = ""
	config.Domainname = ""
	config.ExposedPorts = nil

	hostConfig := *c.inspect.HostConfig

	hostConfig.NetworkMode = container.NetworkMode(containerNetworkMode + rootId)
	hostConfig.PortBindings = nil
	hostConfig.PublishAllPorts = false

	return container_runtime.ContainerCreateRequest{
		Config:        &container_runtime.ContainerConfig{Config: &config},
		HostConfig:    &container_runtime.HostConfig{HostConfig: &hostConfig},
		ContainerName: c.name,
	}
}

func endpointsConfig(inspect container.InspectResponse) map[string]*network.EndpointSettings {
	endpoints := make(map[string]*network.EndpointSettings)

	if inspect.NetworkSettings == nil {
		return endpoints
	}

	for name, endpoint := range inspect.NetworkSettings.Networks {
		if !isUserNetwork(name) {
			continue
		}

		endpoints[name] = &network.EndpointSettings{Aliases: aliasesWithoutId(endpoint, inspect.ID)}
	}

	return endpoints
}

// networkNames lists the user-defined networks a container is attached to.
func networkNames(inspect container.InspectResponse) []string {
	names := make([]string, 0)

	if inspect.NetworkSettings == nil {
		return names
	}

	for name := range inspect.NetworkSettings.Networks {
		if isUserNetwork(name) {
			names = append(names, name)
		}
	}

	slices.Sort(names)

	return names
}

func isUserNetwork(name string) bool {
	switch name {
	case networkBridge, networkHost, networkNone:
		return false
	}

	return !strings.HasPrefix(name, containerNetworkMode)
}

// aliasesWithoutId drops the container id Docker adds as an alias on its own.
func aliasesWithoutId(endpoint *network.EndpointSettings, containerId string) []string {
	if endpoint == nil {
		return nil
	}

	aliases := make([]string, 0, len(endpoint.Aliases))

	for _, alias := range endpoint.Aliases {
		if alias != shortId(containerId) && alias != containerId {
			aliases = append(aliases, alias)
		}
	}

	return aliases
}

func shortId(containerId string) string {
	if len(containerId) <= shortIdLen {
		return containerId
	}

	return containerId[:shortIdLen]
}
