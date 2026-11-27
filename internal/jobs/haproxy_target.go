package jobs

import (
	"context"
	"net"
	"net/netip"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/cluster/env"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service/user_errors"
	"go.vervstack.ru/Velez/internal/storage"
)

const (
	haproxyTargetTimeout = 4 * time.Second
)

func ValidateHaproxyTarget(
	ctx context.Context,
	dataStorage storage.Storage,
	runtimes container_runtime.RuntimeResolver,
	req *velez_api.ValidateHaproxyTarget_Request,
) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, haproxyTargetTimeout)
	defer cancel()

	master := req.GetMasterPgInstance()

	_, err := haproxyMasterUsername(ctx, dataStorage, master)
	if err != nil {
		return "", rerrors.Wrap(err)
	}

	target := req.GetConnectionTarget()

	address, err := haproxyEndpointAddress(target)
	if err != nil {
		return "", rerrors.Wrap(err)
	}

	if target.GetTailscaleEndpoint() != nil {
		return "", user_errors.New("Tailscale reachability can be checked after the sidecar connects")
	}

	if address != "" {
		dialer := net.Dialer{}

		conn, dialErr := dialer.DialContext(ctx, "tcp", address)
		if dialErr != nil {
			return "", rerrors.Wrap(dialErr, "could not reach PostgreSQL over TCP")
		}

		_ = conn.Close()

		return "", nil
	}

	runtime, err := runtimes.Runtime(ctx, req.GetEnvironment())
	if err != nil {
		return "", rerrors.Wrap(err, "error resolving HAProxy container runtime")
	}

	networkName := env.VervNetwork
	if target != nil && target.DockerNetwork != nil {
		networkName = target.GetDockerNetwork()
	}

	shared, err := resolveHaproxyDockerTarget(ctx, runtime, master, networkName)
	if err != nil {
		return "", rerrors.Wrap(err)
	}

	// Docker's bridge IP may be reachable only inside the daemon VM or remote
	// host. Probe the TCP listener inside the PostgreSQL container.
	opts := container.ExecOptions{
		Cmd:          []string{"pg_isready", "-h", shared.ipAddress, "-p", "5432", "-t", "2"},
		AttachStdout: true,
		AttachStderr: true,
	}

	_, exitCode, err := runtime.Exec(ctx, shared.containerId, opts)
	if err != nil {
		return shared.network, rerrors.Wrap(err, "could not check PostgreSQL's TCP listener")
	}

	if exitCode != 0 {
		return shared.network, user_errors.New("PostgreSQL is not accepting connections on the Docker network")
	}

	return shared.network, nil
}

func haproxyEndpointAddress(target *velez_api.HaproxyConnectionTarget) (string, error) {
	public := target.GetPublicEndpoint()
	tailscale := target.GetTailscaleEndpoint()

	if (public != nil && tailscale != nil) ||
		(target.GetDockerNetwork() != "" && (public != nil || tailscale != nil)) {
		return "", user_errors.New("Choose one PostgreSQL connection target")
	}

	endpoint := public
	if tailscale != nil {
		endpoint = tailscale
	}

	if endpoint == nil {
		return "", nil
	}

	host := endpoint.GetHost()
	ip, err := netip.ParseAddr(host)
	isIp := err == nil && ip.Zone() == ""

	if tailscale != nil && !isIp {
		return "", user_errors.New("Enter the PostgreSQL server's Tailscale IP address")
	}

	if !isIp && !haproxyMasterNamePattern.MatchString(host) {
		return "", user_errors.New("Enter a valid PostgreSQL host")
	}

	port := endpoint.GetPort()
	if port == 0 || port > domain.HaproxyMaxPort {
		return "", user_errors.New("Enter a PostgreSQL port between 1 and 65535")
	}

	return net.JoinHostPort(host, strconv.FormatUint(uint64(port), 10)), nil
}

type haproxyDockerTarget struct {
	network     string
	networkId   string
	host        string
	ipAddress   string
	containerId string
}

func resolveHaproxyDockerTarget(
	ctx context.Context,
	runtime container_runtime.ContainerRuntime,
	master, networkName string,
) (haproxyDockerTarget, error) {
	req := &velez_api.ListSmerds_Request{
		Label: map[string]string{labels.VervServiceLabel: master},
	}

	containers, err := runtime.ListContainers(ctx, req)
	if err != nil {
		return haproxyDockerTarget{}, rerrors.Wrap(err, "error listing PostgreSQL containers")
	}

	for _, summary := range containers {
		if summary.State != dockerContainerStatusRunning {
			continue
		}

		info, isFound, inspectErr := runtime.Inspect(ctx, summary.ID)
		if inspectErr != nil {
			return haproxyDockerTarget{}, rerrors.Wrap(inspectErr, "error inspecting PostgreSQL container")
		}

		if !isFound || info.NetworkSettings == nil {
			continue
		}

		target := findHaproxyDockerTarget(ctx, runtime, info, master, networkName)
		if target.network != "" {
			return target, nil
		}
	}

	return haproxyDockerTarget{}, user_errors.New(
		"No running PostgreSQL container shares a Docker network with this deployment")
}

func findHaproxyDockerTarget(
	ctx context.Context,
	runtime container_runtime.ContainerRuntime,
	info container.InspectResponse,
	master, networkName string,
) haproxyDockerTarget {
	var fallback haproxyDockerTarget

	names := make([]string, 0, len(info.NetworkSettings.Networks))

	for name := range info.NetworkSettings.Networks {
		names = append(names, name)
	}

	slices.Sort(names)

	for _, name := range names {
		endpoint := info.NetworkSettings.Networks[name]
		if endpoint == nil || endpoint.NetworkID == "" {
			continue
		}

		ipAddress := endpoint.IPAddress
		if ipAddress == "" {
			ipAddress = endpoint.GlobalIPv6Address
		}

		if ipAddress == "" {
			continue
		}

		networkInfo, err := runtime.InspectNetwork(ctx, endpoint.NetworkID)
		if err != nil || networkInfo.IsSystem || !networkInfo.IsIccEnabled || networkInfo.JoinName == "" {
			continue
		}

		if networkName != "" && networkInfo.JoinName != networkName {
			continue
		}

		host := master
		if !slices.Contains(endpoint.Aliases, master) && !slices.Contains(endpoint.DNSNames, master) {
			host = strings.TrimPrefix(info.Name, "/")
			if len(endpoint.DNSNames) > 0 {
				host = endpoint.DNSNames[0]
			} else if len(endpoint.Aliases) > 0 {
				host = endpoint.Aliases[0]
			}
		}

		if !haproxyMasterNamePattern.MatchString(host) {
			continue
		}

		target := haproxyDockerTarget{
			network: networkInfo.JoinName, networkId: endpoint.NetworkID, host: host, ipAddress: ipAddress, containerId: info.ID,
		}
		if target.network == env.VervNetwork || networkName != "" {
			return target
		}

		if fallback.network == "" {
			fallback = target
		}
	}

	return fallback
}
