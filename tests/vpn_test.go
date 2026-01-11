package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"go.vervstack.ru/Velez/pkg/velez_api"
)

type VpnSuite struct {
	suite.Suite

	env *TestEnvironment

	controlPlaneApi velez_api.ControlPlaneAPIServer
	vpnApi          velez_api.VcnApiServer

	ctx         context.Context
	namespaceId string
	serviceName string
}

func (s *VpnSuite) SetupSuite() {
	//region Declarative preps
	t := s.T()
	t.Parallel()

	s.ctx = s.T().Context()

	s.env = NewEnvironment(t, WithVcn())

	s.controlPlaneApi = s.env.App.Custom.ControlPlaneApiImpl
	s.vpnApi = s.env.App.Custom.VpnApiImpl
	//endregion

	enableVpnRequest := &velez_api.EnableService_Request{
		Service: velez_api.VervServiceType_headscale,
	}
	_, err := s.controlPlaneApi.EnableService(s.ctx, enableVpnRequest)
	require.NoError(s.T(), err, "error enabling vpn service")
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

	_, err := s.env.CreateSmerd(t.Context(), mainApp)
	require.NoError(t, err)

	newNamespaceReq := &velez_api.CreateVcnNamespace_Request{
		Name: s.serviceName,
	}
	newNamespaceResp, err := s.vpnApi.CreateNamespace(t.Context(), newNamespaceReq)
	require.NoError(t, err)

	s.namespaceId = newNamespaceResp.Namespace.Id
}

func (s *VpnSuite) Test_ConnectVpn() {
	t := s.T()

	connectReq := &velez_api.ConnectService_Request{
		ServiceName: s.serviceName,
	}

	connectResp, err := s.vpnApi.ConnectService(t.Context(), connectReq)
	require.NoError(t, err)
	require.NotNil(t, connectResp)
}

func (s *VpnSuite) TearDownTest() {
	if s.namespaceId == "" {
		return
	}

	t := s.T()

	deleteNamespaceReq := &velez_api.DeleteVcnNamespace_Request{
		Id: s.namespaceId,
	}
	_, err := s.vpnApi.DeleteNamespace(t.Context(), deleteNamespaceReq)
	require.NoError(t, err)
}

func Test_Vpn(t *testing.T) {
	suite.Run(t, new(VpnSuite))
}
