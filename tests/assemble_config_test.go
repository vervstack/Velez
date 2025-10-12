package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.redsock.ru/toolbox"

	"go.vervstack.ru/Velez/pkg/velez_api"
)

type AssembleConfigSuite struct {
	suite.Suite

	ctx context.Context
}

func (s *AssembleConfigSuite) SetupSuite() {
	s.ctx = context.Background()
}

func (s *AssembleConfigSuite) Test_AssembleHelloWorld() {
	serviceName := getServiceName(s.Suite.T())
	req := &velez_api.AssembleConfig_Request{
		ImageName:   helloWorldAppImage,
		ServiceName: serviceName,
	}

	assembleResponse, err := testEnvironment.velezAPI.AssembleConfig(s.ctx, req)
	require.NoError(s.Suite.T(), err)

	expected := &velez_api.AssembleConfig_Response{
		Config: s.helloWorldConfig(),
	}

	require.YAMLEq(s.Suite.T(), string(expected.Config), string(assembleResponse.Config))

	listReq := &velez_api.ListSmerds_Request{
		Name: toolbox.ToPtr(serviceName),
	}
	cont, err := testEnvironment.docker.ListContainers(s.ctx, listReq)
	require.NoError(s.Suite.T(), err)
	require.Empty(s.Suite.T(), cont)
}

func (s *AssembleConfigSuite) helloWorldConfig() []byte {
	return []byte(`
app_info:
    name: github.com/godverv/hello_world
    version: v0.0.14
    startup_duration: 10s
data_sources:
    - resource_name: sqlite
      path: ./hello_world.db
      migrations_folder: ./migrations
servers:
    80:
        /{GRPC}:
            module: pkg/hello_world
            gateway: /v1
environment:
    - name: int_slice
      type: int
      value: ['18501:18519']
`)[1:]
}

func Test_AssembleConfig(t *testing.T) {
	suite.Run(t, new(AssembleConfigSuite))
}
