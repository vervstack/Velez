package container_runtime

import (
	"context"
	"net"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	vervNetworkName      = "verv"
	defaultBridgeNetwork = "bridge"
	localhostHost        = "localhost"
)

// Addressing decides how Velez dials a container's port: over a Docker network
// it shares with the container, or through the daemon host's published port.
type Addressing struct {
	ownNetworks   map[string]struct{}
	gatewayIp     string
	isInContainer bool
	dialHost      string
}

// NewAddressing builds the addressing for a daemon reached at daemonHost
// without any shared Docker network.
func NewAddressing(daemonHost string) Addressing {
	return Addressing{dialHost: parseDialHost(daemonHost, false, "")}
}

// DetectAddressing inspects Velez's own container (when it runs in one) to
// learn which Docker networks it shares with the containers it manages. A
// failed inspect is not fatal: it falls back to dialing published ports.
func DetectAddressing(
	ctx context.Context,
	cli client.APIClient,
	daemonHost string,
	ownContainerId *string,
) Addressing {
	if ownContainerId == nil {
		return NewAddressing(daemonHost)
	}

	info, err := cli.ContainerInspect(ctx, *ownContainerId)
	if err != nil {
		log.Warn().Err(err).Msg("error inspecting own container, container addresses use published ports only")

		return Addressing{
			isInContainer: true,
			dialHost:      parseDialHost(daemonHost, true, ""),
		}
	}

	ownNetworks, gatewayIp := ownNetworkSettings(info)

	return Addressing{
		ownNetworks:   ownNetworks,
		gatewayIp:     gatewayIp,
		isInContainer: true,
		dialHost:      parseDialHost(daemonHost, true, gatewayIp),
	}
}

func ownNetworkSettings(info container.InspectResponse) (map[string]struct{}, string) {
	ownNetworks := make(map[string]struct{})
	if info.NetworkSettings == nil {
		return ownNetworks, ""
	}

	gatewayIp := ""

	for name, endpoint := range info.NetworkSettings.Networks {
		ownNetworks[name] = struct{}{}

		if endpoint == nil || endpoint.Gateway == "" {
			continue
		}

		if gatewayIp == "" || name == vervNetworkName {
			gatewayIp = endpoint.Gateway
		}
	}

	return ownNetworks, gatewayIp
}

func parseDialHost(daemonHost string, isInContainer bool, gatewayIp string) string {
	parsed, err := url.Parse(daemonHost)
	if err == nil {
		switch parsed.Scheme {
		case "tcp", "http", "https":
			return parsed.Hostname()
		}
	}

	if isInContainer {
		return gatewayIp
	}

	return localhostHost
}

// Address is the host:port Velez dials to reach containerPort of the inspected
// container.
func (a Addressing) Address(info container.InspectResponse, containerPort int) (string, error) {
	if info.NetworkSettings == nil {
		return "", rerrors.Wrap(user_errors.ErrContainerPortNotPublished)
	}

	sharedHost, isShared := a.sharedNetworkHost(info)
	if isShared {
		return net.JoinHostPort(sharedHost, strconv.Itoa(containerPort)), nil
	}

	bindings := info.NetworkSettings.Ports[nat.Port(strconv.Itoa(containerPort)+"/tcp")]

	hostPort := ""

	for _, binding := range bindings {
		if binding.HostPort != "" {
			hostPort = binding.HostPort

			break
		}
	}

	if hostPort == "" {
		return "", rerrors.Wrap(user_errors.ErrContainerPortNotPublished)
	}

	if a.dialHost == "" {
		return "", rerrors.Wrap(errDaemonHostUnknown)
	}

	return net.JoinHostPort(a.dialHost, hostPort), nil
}

func (a Addressing) sharedNetworkHost(info container.InspectResponse) (string, bool) {
	shared := make([]string, 0, len(info.NetworkSettings.Networks))

	for name, endpoint := range info.NetworkSettings.Networks {
		_, isOwn := a.ownNetworks[name]
		if isOwn && endpoint != nil {
			shared = append(shared, name)
		}
	}

	if len(shared) == 0 {
		return "", false
	}

	slices.Sort(shared)

	chosen := shared[0]
	if slices.Contains(shared, vervNetworkName) {
		chosen = vervNetworkName
	}

	name := strings.TrimPrefix(info.Name, "/")

	// Docker's embedded DNS serves container names on user-defined networks
	// only; the default bridge has no name resolution, so it is dialed by IP.
	ip := info.NetworkSettings.Networks[chosen].IPAddress
	if chosen == defaultBridgeNetwork && ip != "" || name == "" {
		return ip, true
	}

	return name, true
}

// forDaemonHost is the addressing for a dedicated daemon: Velez shares no
// network with it, only its host is dialed.
func (a Addressing) forDaemonHost(daemonHost string) Addressing {
	return Addressing{
		gatewayIp:     a.gatewayIp,
		isInContainer: a.isInContainer,
		dialHost:      parseDialHost(daemonHost, a.isInContainer, a.gatewayIp),
	}
}
