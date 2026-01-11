package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.redsock.ru/toolbox"

	"go.vervstack.ru/Velez/pkg/velez_api"
	"go.vervstack.ru/Velez/tests/config_mocks"
)

type AssembleConfigSuite struct {
	suite.Suite

	ctx context.Context
}

func (s *AssembleConfigSuite) SetupSuite() {
	s.ctx = context.Background()
}

func (s *AssembleConfigSuite) Test_AssembleHelloWorld() {
	t := s.T()

	serviceName := GetServiceName(t)
	req := &velez_api.AssembleConfig_Request{
		ImageName:   HelloWorldAppImage,
		ServiceName: serviceName,
	}

	env := NewEnvironment(t)

	assembleResponse, err := env.App.Custom.ApiGrpcImpl.AssembleConfig(s.ctx, req)
	require.NoError(t, err)

	expected := &velez_api.AssembleConfig_Response{
		Config: config_mocks.HelloWorld,
	}

	require.YAMLEq(t, string(expected.Config), string(assembleResponse.Config))

	listReq := &velez_api.ListSmerds_Request{
		Name: toolbox.ToPtr(serviceName),
	}
	cont, err := env.App.Custom.NodeClients.Docker().ListContainers(s.ctx, listReq)
	require.NoError(t, err)
	require.Empty(t, cont)
}

func Test_AssembleConfig(t *testing.T) {
	suite.Run(t, new(AssembleConfigSuite))
}
