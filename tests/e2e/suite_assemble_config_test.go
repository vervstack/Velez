//go:build e2e_full

package e2e

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.redsock.ru/toolbox"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/tests/config_mocks"
)

type AssembleConfigSuite struct {
	suite.Suite

	plane Plane
	ctx   context.Context
}

func (s *AssembleConfigSuite) SetupSuite() {
	s.ctx = context.Background()
}

func (s *AssembleConfigSuite) Test_AssembleHelloWorld() {
	t := s.T()
	t.Parallel()

	serviceName := GetServiceName(t)

	env := s.plane.NewEnvironment(t)

	req := &velez_api.AssembleConfig_Request{
		ImageName:   HelloWorldAppImage,
		ServiceName: serviceName,
	}

	assembleResponse, err := env.Custom.ApiGrpcImpl.AssembleConfig(s.ctx, req)
	require.NoError(t, err)

	expected := &velez_api.AssembleConfig_Response{
		Config: config_mocks.HelloWorld,
	}

	require.YAMLEq(t, string(expected.GetConfig()), string(assembleResponse.GetConfig()))

	listReq := &velez_api.ListSmerds_Request{
		Name: toolbox.ToPtr(serviceName),
	}
	cont, err := env.Custom.NodeClients.Docker().ListContainers(s.ctx, listReq, "")
	require.NoError(t, err)
	require.Empty(t, cont)
}

func Test_AssembleConfig(t *testing.T) {
	t.Parallel()
	RunPlaneSuite(t, Planes, func(plane Plane) suite.TestingSuite {
		return &AssembleConfigSuite{plane: plane}
	})
}
