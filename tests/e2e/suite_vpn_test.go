package e2e

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

type VpnSuite struct {
	suite.Suite

	env *TestEnvironment

	vpnAPI velez_api.VcnApiClient

	ctx         context.Context
	namespaceID string
	serviceName string
}

func (s *VpnSuite) SetupSuite() {
	t := s.T()

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

	s.vpnAPI = s.env.VpnClient()
}

func (s *VpnSuite) SetupTest() {
	t := s.T()

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
	s.T().Helper()

	t := s.T()
	ctx := t.Context()

	newNamespaceReq := &velez_api.CreateVcnNamespace_Request{
		Name: s.serviceName,
	}

	newNamespaceResp, err := s.vpnAPI.CreateNamespace(ctx, newNamespaceReq)
	if err == nil {
		s.namespaceID = newNamespaceResp.GetNamespace().GetId()

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

	for _, ns := range listNamespacesResp.GetNamespaces() {
		if ns.GetId() == s.serviceName {
			s.namespaceID = ns.GetId()
		}
	}
}

func Test_Vpn(t *testing.T) {
	t.Skip("")
	suite.Run(t, new(VpnSuite))
}
