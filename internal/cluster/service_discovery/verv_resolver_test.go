package service_discovery

import (
	"context"
	"net"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	"go.vervstack.ru/makosh/pkg/makosh_be"
	matreshkaSd "go.vervstack.ru/matreshka/pkg/matreshka/service_discovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/resolver"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
)

const (
	unreachableAddr = "127.0.0.1:1"
	discoveryAddr   = "10.0.0.1:1"
)

type inProcessDiscovery struct {
	cluster_clients.ServiceDiscovery

	endpoints map[string][]string
}

func (d *inProcessDiscovery) ListEndpoints(
	_ context.Context,
	req *makosh_be.ListEndpoints_Request,
	_ ...grpc.CallOption,
) (*makosh_be.ListEndpoints_Response, error) {
	resp := &makosh_be.ListEndpoints_Response{Urls: d.endpoints[req.GetServiceName()]}

	return resp, nil
}

type recordingClientConn struct {
	resolver.ClientConn

	state resolver.State
}

func (c *recordingClientConn) UpdateState(state resolver.State) error {
	c.state = state

	return nil
}

func (c *recordingClientConn) ReportError(error) {}

func newOverrides(serviceName string, urls ...string) matreshkaSd.Overrides {
	override := &matreshkaSd.Override{ServiceName: serviceName, Urls: urls}

	return matreshkaSd.Overrides{override}
}

func resolveAddrs(t *testing.T, builder resolver.Builder, serviceName string) []string {
	t.Helper()

	cc := &recordingClientConn{}

	parsed, err := url.Parse("verv://" + serviceName)
	require.NoError(t, err)

	target := resolver.Target{URL: *parsed}

	r, buildErr := builder.Build(target, cc, resolver.BuildOptions{})
	require.NoError(t, buildErr)

	defer r.Close()

	addrs := make([]string, 0, len(cc.state.Addresses))
	for _, a := range cc.state.Addresses {
		addrs = append(addrs, a.Addr)
	}

	return addrs
}

func Test_VervResolver_OverrideWinsOverDiscovery(t *testing.T) {
	discovery := &inProcessDiscovery{endpoints: map[string][]string{"matreshka": {discoveryAddr}}}
	builder := NewVervResolverBuilder(newOverrides("matreshka", "192.168.0.1:2"), discovery)

	require.Equal(t, []string{"192.168.0.1:2"}, resolveAddrs(t, builder, "matreshka"))
}

func Test_VervResolver_FallsBackToDiscoveryWithoutOverride(t *testing.T) {
	discovery := &inProcessDiscovery{endpoints: map[string][]string{"matreshka": {discoveryAddr}}}
	builder := NewVervResolverBuilder(newOverrides("other", "192.168.0.1:2"), discovery)

	require.Equal(t, []string{discoveryAddr}, resolveAddrs(t, builder, "matreshka"))
}

func Test_VervResolver_UnknownServiceResolvesNothing(t *testing.T) {
	builder := NewVervResolverBuilder(nil, &inProcessDiscovery{})

	require.Empty(t, resolveAddrs(t, builder, "matreshka"))
}

func Test_VervResolver_BuildersAreIsolatedPerApp(t *testing.T) {
	first := NewVervResolverBuilder(newOverrides("matreshka", "1.1.1.1:1"), &inProcessDiscovery{})
	second := NewVervResolverBuilder(newOverrides("matreshka", "2.2.2.2:2"), &inProcessDiscovery{})

	require.Equal(t, []string{"1.1.1.1:1"}, resolveAddrs(t, first, "matreshka"))
	require.Equal(t, []string{"2.2.2.2:2"}, resolveAddrs(t, second, "matreshka"))
}

func Test_VervResolver_GrpcClientReachesServerBehindOverride(t *testing.T) {
	var listenCfg net.ListenConfig

	lis, err := listenCfg.Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)

	server := grpc.NewServer()
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())

	go func() { _ = server.Serve(lis) }()

	t.Cleanup(server.Stop)

	discovery := &inProcessDiscovery{endpoints: map[string][]string{"svc": {unreachableAddr}}}
	builder := NewVervResolverBuilder(newOverrides("svc", lis.Addr().String()), discovery)

	conn, err := grpc.NewClient("verv://svc",
		grpc.WithResolvers(builder),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	t.Cleanup(func() { _ = conn.Close() })

	req := &grpc_health_v1.HealthCheckRequest{}

	_, err = grpc_health_v1.NewHealthClient(conn).Check(t.Context(), req)
	require.NoError(t, err)
}
