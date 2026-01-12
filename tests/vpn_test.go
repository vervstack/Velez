package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.vervstack.ru/Velez/pkg/velez_api"
)

type VpnSuite struct {
	suite.Suite

	env *TestEnvironment

	controlPlaneApi velez_api.ControlPlaneAPIServer
	vpnApi          velez_api.VcnApiClient

	ctx         context.Context
	namespaceId string
	serviceName string
}

func (s *VpnSuite) SetupSuite() {
	//region Declarative preps
	t := s.T()
	t.Parallel()

	s.ctx = s.T().Context()

	s.env = NewEnvironment(t,
		WithState(t,
			WithStateVcnEnabled()))

	s.controlPlaneApi = s.env.App.Custom.ControlPlaneApiImpl
	s.vpnApi = s.env.VpnClient()
	//endregion
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

	s.prepareNamespace()
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

	// TODO not only delete namespace but also drop

	t := s.T()

	r := &velez_api.DeleteVcnNamespace_Request{
		Id: s.namespaceId,
	}
	_, err := s.vpnApi.DeleteNamespace(s.ctx, r)
	require.NoError(t, err)
}

func (s *VpnSuite) prepareNamespace() {
	t := s.T()
	ctx := t.Context()
	newNamespaceReq := &velez_api.CreateVcnNamespace_Request{
		Name: s.serviceName,
	}

	newNamespaceResp, err := s.vpnApi.CreateNamespace(ctx, newNamespaceReq)
	if err == nil {
		s.namespaceId = newNamespaceResp.Namespace.Id
		return
	}

	c, k := status.FromError(err)
	if !(k && c.Code() == codes.AlreadyExists) {
		require.NoError(t, err)
	}
	// TODO add listing with name
	listReq := &velez_api.ListVcnNamespaces_Request{}

	listNamespacesResp, err := s.vpnApi.ListNamespaces(ctx, listReq)
	require.NoError(t, err)

	for _, ns := range listNamespacesResp.Namespaces {
		if ns.Id == s.serviceName {
			s.namespaceId = ns.Id
		}
	}
}

func Test_Vpn(t *testing.T) {
	suite.Run(t, new(VpnSuite))
}
