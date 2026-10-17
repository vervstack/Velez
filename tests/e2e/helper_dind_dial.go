package e2e

import (
	"context"
	"net"
	"net/http"
	"slices"
	"strconv"
	"sync"
	"testing"

	dockernet "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"
	"golang.org/x/net/proxy"

	clusterenv "go.vervstack.ru/Velez/internal/cluster/env"
)

var (
	//nolint:forbidigo // package-private test-infra sentinel, not shared/user-facing
	errDindNoContainerIp = rerrors.New("container has no ip on any network")
	//nolint:forbidigo // package-private test-infra sentinel, not shared/user-facing
	errDindNotContextDialer = rerrors.New("socks5 dialer is not a context dialer")
	//nolint:forbidigo // package-private test-infra sentinel, not shared/user-facing
	errDindNoSocksProxy = rerrors.New("dind has no socks5 proxy; run the suite through tests/e2egroups")

	dindGatewayOnce sync.Once //nolint:gochecknoglobals // one gateway per daemon
	dindGatewayIp   string    //nolint:gochecknoglobals // outcome of the one-time lookup
	errDindGateway  error     //nolint:gochecknoglobals // outcome of the one-time lookup
)

// dindGateway is the verv network's gateway: the address every port bound
// inside the DinD answers at, both for containers on that network and, through
// the SOCKS5 proxy, for the test process.
func dindGateway() (string, error) {
	dindGatewayOnce.Do(func() {
		dockerClient, err := client.NewClientWithOpts(
			client.WithHost(sharedDind.DockerHost), client.WithAPIVersionNegotiation())
		if err != nil {
			errDindGateway = rerrors.Wrap(err, "error creating docker client for the dind gateway")

			return
		}

		defer func() { _ = dockerClient.Close() }()

		inspectOpts := dockernet.InspectOptions{}

		inspected, err := dockerClient.NetworkInspect(context.Background(), clusterenv.VervNetwork, inspectOpts)
		if err != nil {
			errDindGateway = rerrors.Wrap(err, "error inspecting the verv network")

			return
		}

		if len(inspected.IPAM.Config) == 0 {
			errDindGateway = rerrors.Wrap(errDindNoContainerIp, "verv network has no ipam config")

			return
		}

		dindGatewayIp = inspected.IPAM.Config[0].Gateway
	})

	return dindGatewayIp, errDindGateway
}

// dindGatewayAddr is gateway:port for a port a fixture or a container publishes
// inside the DinD.
func dindGatewayAddr(t *testing.T, port int) string {
	t.Helper()

	gateway, err := dindGateway()
	require.NoError(t, err)

	return net.JoinHostPort(gateway, strconv.Itoa(port))
}

// dindDialIp dials an ip:port address inside the DinD through the group's
// SOCKS5 proxy.
func dindDialIp(ctx context.Context, network, addr string) (net.Conn, error) {
	socksAddr, ok := sharedDind.SocksAddr()
	if !ok {
		return nil, errDindNoSocksProxy
	}

	socksDialer, err := proxy.SOCKS5("tcp", socksAddr, nil, &net.Dialer{})
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating socks5 dialer")
	}

	contextDialer, ok := socksDialer.(proxy.ContextDialer)
	if !ok {
		return nil, errDindNotContextDialer
	}

	conn, err := contextDialer.DialContext(ctx, network, addr)
	if err != nil {
		return nil, rerrors.Wrap(err, "error dialing through socks5 proxy")
	}

	return conn, nil
}

func dindDialIpAddr(ctx context.Context, addr string) (net.Conn, error) {
	return dindDialIp(ctx, "tcp", addr)
}

// DindDialContext dials addr from the test process. Container names and
// DinD-internal IPs go through the group's SOCKS5 proxy, so they resolve the
// way they do for a containerized Velez.
func (e *TestEnvironment) DindDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, rerrors.Wrap(err, "error splitting dial address "+addr)
	}

	ip, err := e.dindHostIp(ctx, host)
	if err != nil {
		return nil, rerrors.Wrap(err, "error resolving dind host "+host)
	}

	return dindDialIp(ctx, network, net.JoinHostPort(ip, port))
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
