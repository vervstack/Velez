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

	controlPlaneApi velez_api.ControlPlaneAPIClient
	vpnApi          velez_api.VpnApiClient

	ctx         context.Context
	namespaceId string
	serviceName string
}

func (s *VpnSuite) SetupSuite() {
	//region Declarative preps
	s.ctx = s.T().Context()

	s.controlPlaneApi = testEnvironment.api.controlPlane
	s.vpnApi = testEnvironment.api.vpn

	//endregion

	//region Imperative preps
	enableVpnRequest := &velez_api.EnableServices_Request{
		Services: []velez_api.VervServiceType{
			velez_api.VervServiceType_headscale,
		},
	}
	_, err := s.controlPlaneApi.EnableServices(s.ctx, enableVpnRequest)
	require.NoError(s.T(), err, "error enabling vpn service")
	//endregion
}

func (s *VpnSuite) SetupTest() {
	s.serviceName = getServiceName(s.T())
	mainApp := &velez_api.CreateSmerd_Request{
		Name:         s.serviceName,
		ImageName:    helloWorldAppImage,
		IgnoreConfig: true,
	}

	_, err := testEnvironment.createSmerd(s.T().Context(), mainApp)
	require.NoError(s.T(), err)

	newNamespaceReq := &velez_api.CreateVpnNamespace_Request{
		Name: s.serviceName,
	}
	newNamespaceResp, err := s.vpnApi.CreateNamespace(s.T().Context(), newNamespaceReq)
	require.NoError(s.T(), err)

	s.namespaceId = newNamespaceResp.Namespace.Id
}

func (s *VpnSuite) Test_ConnectVpn() {
	connectReq := &velez_api.ConnectService_Request{
		ServiceName: s.serviceName,
	}

	connectResp, err := s.vpnApi.ConnectService(s.T().Context(), connectReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), connectResp)
}

func (s *VpnSuite) TearDownTest() {
	if s.namespaceId == "" {
		return
	}

	deleteNamespaceReq := &velez_api.DeleteVpnNamespace_Request{
		Id: s.namespaceId,
	}
	_, err := s.vpnApi.DeleteNamespace(s.T().Context(), deleteNamespaceReq)
	require.NoError(s.T(), err)
}

func Test_Vpn(t *testing.T) {
	suite.Run(t, new(VpnSuite))
}
