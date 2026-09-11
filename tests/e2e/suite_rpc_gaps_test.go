//go:build e2e_full

package e2e

import (
	"context"
	"testing"

	dockernetwork "github.com/docker/docker/api/types/network"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

// RpcGapsSuite is thin "is it reachable and shaped right" coverage for RPCs
// that had no e2e touch before: the pure getters on the velez_api impl
// (Version, SearchImages, GetHardware), the service-read RPCs on the
// service_api impl (GetServiceMetrics/Resources/Graph/Environments and
// GetVervonomicon), and a docker-network round-trip through MakeConnections /
// BreakConnections. None of these need cluster Postgres or matreshka, so
// every test runs a plain NewEnvironment(t). GetVervonomicon's real behavior
// (image-sourced descriptors, environment overlays, resource reconciliation,
// error cases) is covered in suite_vervonomicon_test.go and
// suite_vervonomicon_deploy_test.go - this suite only proves the unknown-
// service shape (a clean empty response).
type RpcGapsSuite struct {
	suite.Suite

	ctx context.Context
}

func (s *RpcGapsSuite) SetupSuite() {
	s.ctx = context.Background()
}

// Test_Version is a pure getter: the version string is seeded from
// cfg.AppInfo.Version by NewEnvironment, so it must come back non-empty.
func (s *RpcGapsSuite) Test_Version() {
	t := s.T()

	env := NewEnvironment(t)

	resp, err := env.Custom.ApiGrpcImpl.Version(s.ctx, &velez_api.Version_Request{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.GetVersion())
}

// Test_SearchImages exercises the handler's reachability only.
//
// NOTE: image_list.go builds &velez_api.SearchImages_Response{Images:
// []*velez_api.SearchImageItem{}} unconditionally - the result slice is
// always empty regardless of the query - so the only thing observable here
// is whether the underlying dockerutils.SearchImages call errored. Do NOT
// assert non-empty images.
func (s *RpcGapsSuite) Test_SearchImages() {
	t := s.T()

	env := NewEnvironment(t)

	req := &velez_api.SearchImages_Request{Name: "nginx"}

	resp, err := env.Custom.ApiGrpcImpl.SearchImages(s.ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// Test_GetHardware asserts only the host-independent shape: no error and the
// three sub-messages (cpu / ram / disk) are always allocated by the hardware
// manager, even when a probe fails (the failure lands in Value.Err, not a
// nil message). Actual CPU/RAM/disk strings are machine-dependent and not
// asserted.
func (s *RpcGapsSuite) Test_GetHardware() {
	t := s.T()

	env := NewEnvironment(t)

	resp, err := env.Custom.ApiGrpcImpl.GetHardware(s.ctx, &velez_api.GetHardware_Request{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.GetCpu())
	require.NotNil(t, resp.GetRam())
	require.NotNil(t, resp.GetDiskMem())
}

// Test_GetServiceMetrics: a service name that was never deployed resolves to
// zero smerds, so the handler returns an empty metrics response, not an
// error.
func (s *RpcGapsSuite) Test_GetServiceMetrics() {
	t := s.T()

	env := NewEnvironment(t)

	req := &velez_api.GetServiceMetrics_Request{ServiceName: GetServiceName(t)}

	resp, err := env.Custom.ServiceApiImpl.GetServiceMetrics(s.ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Zero(t, resp.GetReplicasDesired())
}

// Test_GetServiceResources: unknown service -> empty (non-nil) resources,
// no error.
func (s *RpcGapsSuite) Test_GetServiceResources() {
	t := s.T()

	env := NewEnvironment(t)

	req := &velez_api.GetServiceResources_Request{ServiceName: GetServiceName(t)}

	resp, err := env.Custom.ServiceApiImpl.GetServiceResources(s.ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Empty(t, resp.GetResources())
}

// Test_GetServiceGraph: unknown service -> empty dependencies and callers,
// no error.
func (s *RpcGapsSuite) Test_GetServiceGraph() {
	t := s.T()

	env := NewEnvironment(t)

	req := &velez_api.GetServiceGraph_Request{ServiceName: GetServiceName(t)}

	resp, err := env.Custom.ServiceApiImpl.GetServiceGraph(s.ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Empty(t, resp.GetDependencies())
	require.Empty(t, resp.GetCallers())
}

// Test_GetServiceEnvironments: unknown service -> empty environments,
// no error.
func (s *RpcGapsSuite) Test_GetServiceEnvironments() {
	t := s.T()

	env := NewEnvironment(t)

	req := &velez_api.GetServiceEnvironments_Request{ServiceName: GetServiceName(t)}

	resp, err := env.Custom.ServiceApiImpl.GetServiceEnvironments(s.ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Empty(t, resp.GetEnvironments())
}

// Test_GetVervonomicon: a service name that was never deployed has no image
// to read a descriptor from, so GetVervonomicon returns a clean empty
// response, not an error - the same NoDescriptor path a real service with no
// running instance takes. See suite_vervonomicon_test.go for descriptor
// content, environment overlays, and resource reconciliation coverage.
func (s *RpcGapsSuite) Test_GetVervonomicon() {
	t := s.T()

	env := NewEnvironment(t)

	resp, err := env.Custom.ServiceApiImpl.GetVervonomicon(s.ctx, &velez_api.GetVervonomicon_Request{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// Test_MakeAndBreakConnections is a live docker-network round-trip:
// MakeConnections / BreakConnections route through
// smerdService.ConnectToNetwork / DisconnectFromNetwork, which are plain
// docker network attach/detach ops (no headscale), so they run inside the
// DinD.
//
// The Connection.ServiceName field is passed the smerd UUID: the handler's
// toConnection copies it into domain.Connection.SmerdName, which the runtime
// resolves via resolveOwnedContainer - that accepts a raw docker id as well
// as a logical name. The target network name round-trips unchanged because
// the e2e environment runs with an empty container suffix
// (resolver.NetworkName is the identity).
func (s *RpcGapsSuite) Test_MakeAndBreakConnections() {
	t := s.T()

	env := NewEnvironment(t)
	ctx := t.Context()

	dockerClient := env.Custom.NodeClients.Docker().Client()

	smerdReq := &velez_api.CreateSmerd_Request{
		Name:         GetServiceName(t),
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
	}

	smerd := env.CreateSmerd(t, smerdReq)
	require.NotEmpty(t, smerd.GetUuid())

	netName := GetServiceName(t) + "_net"

	createNetOpts := dockernetwork.CreateOptions{Driver: "bridge"}

	_, err := dockerClient.NetworkCreate(ctx, netName, createNetOpts)
	require.NoError(t, err)

	t.Cleanup(func() {
		cleanupCtx := context.Background()

		_ = dockerClient.NetworkDisconnect(cleanupCtx, netName, smerd.GetUuid(), true)
		_ = dockerClient.NetworkRemove(cleanupCtx, netName)
	})

	conn := &velez_api.Connection{
		ServiceName:   smerd.GetUuid(),
		TargetNetwork: netName,
	}

	makeReq := &velez_api.MakeConnections_Request{
		Connections: []*velez_api.Connection{conn},
	}

	_, err = env.Custom.ApiGrpcImpl.MakeConnections(ctx, makeReq)
	require.NoError(t, err)

	info, err := dockerClient.ContainerInspect(ctx, smerd.GetUuid())
	require.NoError(t, err)
	require.Contains(t, info.NetworkSettings.Networks, netName,
		"the target network must be attached after MakeConnections")

	breakReq := &velez_api.BreakConnections_Request{
		Connections: []*velez_api.Connection{conn},
	}

	_, err = env.Custom.ApiGrpcImpl.BreakConnections(ctx, breakReq)
	require.NoError(t, err)

	info, err = dockerClient.ContainerInspect(ctx, smerd.GetUuid())
	require.NoError(t, err)
	require.NotContains(t, info.NetworkSettings.Networks, netName,
		"the target network must be detached after BreakConnections")
}

// Test_RpcGaps deliberately does NOT call t.Parallel(): its subtests are
// cheap RPC reachability checks, but Test_MakeAndBreakConnections creates a
// container + a docker network, and running that concurrently with the
// parallel deploy suites adds enough docker-daemon load to intermittently
// trip the pre-existing Test_Negative_HealthcheckNeverHealthy race (product
// gap G2). Running this suite serially keeps that suite stable at no
// meaningful wall-clock cost (~7s).
func Test_RpcGaps(t *testing.T) {
	suite.Run(t, new(RpcGapsSuite))
}
