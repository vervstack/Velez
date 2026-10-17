package e2e

import (
	"context"
	"net"
	"net/http"
	"slices"

	"go.redsock.ru/rerrors"
	"golang.org/x/net/proxy"

	clusterenv "go.vervstack.ru/Velez/internal/cluster/env"
)

const (
	dindLocalhostName = "localhost"
)

var (
	//nolint:forbidigo // package-private test-infra sentinel, not shared/user-facing
	errDindNoContainerIp = rerrors.New("container has no ip on any network")
	//nolint:forbidigo // package-private test-infra sentinel, not shared/user-facing
	errDindNotContextDialer = rerrors.New("socks5 dialer is not a context dialer")
)

// DindDialContext dials addr from the test process. Container names and
// DinD-internal IPs go through the group's SOCKS5 proxy, so they resolve the
// way they do for a containerized Velez. localhost / 127.0.0.1 are the
// addresses an in-process Velez hands out; they stay a plain dial (reachable
// through the loopback bridge).
func (e *TestEnvironment) DindDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	direct := &net.Dialer{}

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, rerrors.Wrap(err, "error splitting dial address "+addr)
	}

	socksAddr, ok := sharedDind.SocksAddr()
	if !ok || host == dindLocalhostName || isLoopbackHost(host) {
		return dialDirect(ctx, direct, network, addr)
	}

	ip, err := e.dindHostIp(ctx, host)
	if err != nil {
		return nil, rerrors.Wrap(err, "error resolving dind host "+host)
	}

	socksDialer, err := proxy.SOCKS5("tcp", socksAddr, nil, direct)
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating socks5 dialer")
	}

	contextDialer, ok := socksDialer.(proxy.ContextDialer)
	if !ok {
		return nil, errDindNotContextDialer
	}

	conn, err := contextDialer.DialContext(ctx, network, net.JoinHostPort(ip, port))
	if err != nil {
		return nil, rerrors.Wrap(err, "error dialing through socks5 proxy")
	}

	return conn, nil
}

func dialDirect(ctx context.Context, dialer *net.Dialer, network, addr string) (net.Conn, error) {
	conn, err := dialer.DialContext(ctx, network, addr)
	if err != nil {
		return nil, rerrors.Wrap(err, "error dialing "+addr)
	}

	return conn, nil
}

// DindHttpClient is an http client whose connections go through
// DindDialContext. Its idle connections are closed when the test ends.
func (e *TestEnvironment) DindHttpClient() *http.Client {
	transport := e.DindTransport()

	return &http.Client{Transport: transport}
}

// DindTransport is the transport behind DindHttpClient, for clients that take
// one (minio).
func (e *TestEnvironment) DindTransport() *http.Transport {
	transport := &http.Transport{DialContext: e.DindDialContext}

	e.t.Cleanup(transport.CloseIdleConnections)

	return transport
}

func isLoopbackHost(host string) bool {
	ip := net.ParseIP(host)

	return ip != nil && ip.IsLoopback()
}

func (e *TestEnvironment) dindHostIp(ctx context.Context, host string) (string, error) {
	if net.ParseIP(host) != nil {
		return host, nil
	}

	inspected, err := e.DockerClient().ContainerInspect(ctx, host)
	if err != nil {
		return "", rerrors.Wrap(err, "error inspecting container "+host)
	}

	if inspected.NetworkSettings == nil {
		return "", errDindNoContainerIp
	}

	preferred, ok := inspected.NetworkSettings.Networks[clusterenv.VervNetwork]
	if ok && preferred.IPAddress != "" {
		return preferred.IPAddress, nil
	}

	names := make([]string, 0, len(inspected.NetworkSettings.Networks))

	for name := range inspected.NetworkSettings.Networks {
		names = append(names, name)
	}

	slices.Sort(names)

	for _, name := range names {
		ip := inspected.NetworkSettings.Networks[name].IPAddress
		if ip != "" {
			return ip, nil
		}
	}

	return "", rerrors.Wrap(errDindNoContainerIp, host)
}
