//go:build e2e_full

package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/patterns"
)

// VpnSuite drives the VcnApi RPCs against a real headscale (see
// shared_headscale.go for the fixture and the scope caveat):
//
//   - namespace CRUD  -> real headscale users + real sqlite, end to end
//     through transport -> verv_closed_network client -> headscale HTTP API.
//   - ConnectService  -> the connect_service_to_vpn task really creates a
//     namespace and issues a pre-auth key on the fixture headscale, then
//     really creates and starts a tailscale sidecar container. The sidecar's
//     tailnet join is NOT asserted: getLoginServerURLJob hardcodes
//     https://vcn.redsock.ru, so a hermetic join is not testable until that
//     URL is configurable (Trello #126 follow-up).
//
// Not t.Parallel(): the suite shares one headscale container and asserts on
// its global namespace listing.
type VpnSuite struct {
	suite.Suite

	plane  Plane
	env    *TestEnvironment
	vpnAPI velez_api.VcnApiClient
}

func (s *VpnSuite) SetupSuite() {
	t := s.T()

	hs := getSharedHeadscale(t)

	s.env = s.plane.NewEnvironment(t,
		WithState(t, WithStateVcnEnabled(hs.apiURL, hs.apiKey)))

	s.vpnAPI = s.env.VpnClient()
}

func (s *VpnSuite) Test_NamespaceCrud() {
	t := s.T()
	ctx := t.Context()

	name := GetServiceName(t)

	createReq := &velez_api.CreateVcnNamespace_Request{Name: name}

	created, err := s.vpnAPI.CreateNamespace(ctx, createReq)
	require.NoError(t, err)
	require.NotEmpty(t, created.GetNamespace().GetId(), "a created namespace must carry an id")

	t.Cleanup(func() {
		delReq := &velez_api.DeleteVcnNamespace_Request{Id: created.GetNamespace().GetId()}

		_, _ = s.vpnAPI.DeleteNamespace(context.Background(), delReq)
	})

	require.Eventually(t, func() bool {
		return s.namespacePresent(ctx, name)
	}, 10*time.Second, time.Second, "the created namespace must show up in ListNamespaces")

	_, err = s.vpnAPI.CreateNamespace(ctx, createReq)
	require.Error(t, err, "creating the same namespace twice must fail")
	require.Equal(t, codes.AlreadyExists, status.Code(err), "the duplicate must map to AlreadyExists")

	delReq := &velez_api.DeleteVcnNamespace_Request{Id: created.GetNamespace().GetId()}

	_, err = s.vpnAPI.DeleteNamespace(ctx, delReq)
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		return !s.namespacePresent(ctx, name)
	}, 10*time.Second, time.Second, "the deleted namespace must disappear from ListNamespaces")
}

func (s *VpnSuite) Test_ConnectService_LaunchesSidecar() {
	t := s.T()
	ctx := t.Context()

	serviceName := GetServiceName(t)

	smerd := newHelloWorldRequest(serviceName)
	s.env.CreateSmerd(t, smerd)

	t.Cleanup(func() {
		id := s.findNamespaceID(context.Background(), serviceName)
		if id == "" {
			return
		}

		delReq := &velez_api.DeleteVcnNamespace_Request{Id: id}

		_, _ = s.vpnAPI.DeleteNamespace(context.Background(), delReq)
	})

	connectReq := &velez_api.ConnectService_Request{ServiceName: serviceName}

	_, err := s.vpnAPI.ConnectService(ctx, connectReq)
	require.NoError(t, err, "connect_service_to_vpn must reach DONE against the real headscale")

	// The task issued a real pre-auth key, so the namespace exists on headscale.
	require.True(t, s.namespacePresent(ctx, serviceName),
		"ConnectService must have created the service's namespace on headscale")

	// And the sidecar container is really up (tailnet join not asserted).
	sidecarName := serviceName + "-" + patterns.TailscaleSidecarSuffix

	inspected, err := s.env.Custom.NodeClients.Docker().Client().ContainerInspect(ctx, sidecarName)
	require.NoError(t, err, "the tailscale sidecar container %q must exist", sidecarName)
	require.NotNil(t, inspected.State)
	require.True(t, inspected.State.Running, "the tailscale sidecar container must be running")
}

func (s *VpnSuite) namespacePresent(ctx context.Context, name string) bool {
	return s.findNamespaceID(ctx, name) != ""
}

func (s *VpnSuite) findNamespaceID(ctx context.Context, name string) string {
	t := s.T()

	resp, err := s.vpnAPI.ListNamespaces(ctx, &velez_api.ListVcnNamespaces_Request{})
	require.NoError(t, err)

	for _, ns := range resp.GetNamespaces() {
		if ns.GetName() == name {
			return ns.GetId()
		}
	}

	return ""
}

func Test_Vpn(t *testing.T) {
	RunPlaneSuite(t, Planes, func(plane Plane) suite.TestingSuite {
		return &VpnSuite{plane: plane}
	})
}
