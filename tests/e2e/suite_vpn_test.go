package e2e

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type VpnSuite struct {
	suite.Suite

	env *TestEnvironment

	controlPlaneAPI velez_api.ControlPlaneAPIServer
	vpnAPI          velez_api.VcnApiClient

	ctx         context.Context
	namespaceID string
	serviceName string
}

func (s *VpnSuite) SetupSuite() {
	t := s.T()
	t.Parallel()

	dialer := &net.Dialer{Timeout: 2 * time.Second}

	conn, err := dialer.DialContext(t.Context(), "tcp", "localhost:8080")
	if err != nil {
		t.Skip("headscale not available at localhost:8080, skipping VPN tests")
	}

	err = conn.Close()
	require.NoError(t, err)

	s.ctx = t.Context()

	s.env = NewEnvironment(t,
		WithState(t,
			WithStateVcnEnabled()))

	s.controlPlaneAPI = s.env.Custom.ControlPlaneApiImpl
	s.vpnAPI = s.env.VpnClient()
}

func (s *VpnSuite) SetupTest() {
	t := s.T()
	t.Parallel()

	s.serviceName = GetServiceName(t)

	mainApp := &velez_api.CreateSmerd_Request{
		Name:         s.serviceName,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
	}

	s.env.CreateSmerd(t, mainApp)

	s.prepareNamespace()
}

func (s *VpnSuite) Test_ConnectVpn() {
	t := s.T()

	connectReq := &velez_api.ConnectService_Request{
		ServiceName: s.serviceName,
	}

	connectResp, err := s.vpnAPI.ConnectService(t.Context(), connectReq)
	require.NoError(t, err)
	require.NotNil(t, connectResp)
}

func (s *VpnSuite) TearDownTest() {
	if s.namespaceID == "" {
		return
	}

	t := s.T()

	r := &velez_api.DeleteVcnNamespace_Request{
		Id: s.namespaceID,
	}
	_, err := s.vpnAPI.DeleteNamespace(s.ctx, r)
	assert.NoError(t, err)
}

func (s *VpnSuite) prepareNamespace() {
	t := s.T()
	ctx := t.Context()

	newNamespaceReq := &velez_api.CreateVcnNamespace_Request{
		Name: s.serviceName,
	}

	newNamespaceResp, err := s.vpnAPI.CreateNamespace(ctx, newNamespaceReq)
	if err == nil {
		s.namespaceID = newNamespaceResp.Namespace.Id

		return
	}

	c, k := status.FromError(err)
	if !k || c.Code() != codes.AlreadyExists {
		require.NoError(t, err)
	}

	// TODO add listing with name
	listReq := &velez_api.ListVcnNamespaces_Request{}

	listNamespacesResp, err := s.vpnAPI.ListNamespaces(ctx, listReq)
	require.NoError(t, err)

	for _, ns := range listNamespacesResp.Namespaces {
		if ns.Id == s.serviceName {
			s.namespaceID = ns.Id
		}
	}
}

func Test_Vpn(t *testing.T) {
	suite.Run(t, new(VpnSuite))
}
