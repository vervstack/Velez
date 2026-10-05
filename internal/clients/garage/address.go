package garage

import (
	"context"
	"net"
	"strconv"

	"github.com/docker/docker/api/types/container"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/cluster/env"
	"go.vervstack.ru/Velez/internal/domain"
)

const (
	adminPortKey = "3903/tcp"
	hostLoopback = "localhost"
	httpScheme   = "http://"
)

// AdminUrl resolves the address Velez reaches the admin API of the Garage
// container on.
func AdminUrl(
	ctx context.Context,
	runtime container_runtime.ContainerRuntime,
	containerName string,
) (string, error) {
	info, isFound, err := runtime.Inspect(ctx, containerName)
	if err != nil {
		return "", rerrors.Wrap(err, "error inspecting garage container")
	}

	if !isFound {
		return "", rerrors.Wrap(errContainerNotFound)
	}

	if !env.IsInContainer() {
		return hostAdminUrl(info)
	}

	return networkAdminUrl(info, containerName)
}

func hostAdminUrl(info container.InspectResponse) (string, error) {
	if info.NetworkSettings == nil {
		return "", rerrors.Wrap(errAdminPortNotPublished)
	}

	for port, bindings := range info.NetworkSettings.Ports {
		if string(port) != adminPortKey {
			continue
		}

		for _, binding := range bindings {
			if binding.HostPort != "" {
				return httpScheme + net.JoinHostPort(hostLoopback, binding.HostPort), nil
			}
		}
	}

	return "", rerrors.Wrap(errAdminPortNotPublished)
}

func networkAdminUrl(info container.InspectResponse, containerName string) (string, error) {
	if info.NetworkSettings == nil {
		return "", rerrors.Wrap(errNotOnVervNetwork)
	}

	vervNetwork, isAttached := info.NetworkSettings.Networks[env.VervNetwork]
	if !isAttached || vervNetwork == nil {
		return "", rerrors.Wrap(errNotOnVervNetwork)
	}

	host := containerName
	if len(vervNetwork.Aliases) > 0 {
		host = vervNetwork.Aliases[0]
	}

	port := strconv.Itoa(domain.S3AdminContainerPort)

	return httpScheme + net.JoinHostPort(host, port), nil
}
