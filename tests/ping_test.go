package tests

import (
	"context"
	"testing"

	"github.com/godverv/matreshka-be/pkg/matreshka_api"
	"github.com/stretchr/testify/suite"

	"github.com/godverv/Velez/tests/setup"
)

type PingSuite struct {
	suite.Suite

	ctx context.Context
}

func (s *PingSuite) SetupTest() {
	a, err := setup.GetApp()
	s.Require().NoError(err)

	s.ctx = context.Background()

	c, err := a.Matreshka.UpdateServiceConfig(s.ctx,
		&matreshka_api.UpdateServiceConfig_Request{
			Config: &matreshka_api.Config{
				AppConfig: &matreshka_api.Config_AppConfig{
					Name:            "Red-Sock/Red-Cart",
					Version:         "v0.0.8",
					StartupDuration: "10s",
				},
				Api: []*matreshka_api.Config_Api{
					{
						ApiType: matreshka_api.Config_Api_grpc,
					},
				},
				Resources: []*matreshka_api.Config_Resource{
					{
						ResourceType: matreshka_api.Config_Resource_postgres,
					},
				},
			},
		})

	s.Require().NoError(err)
	_ = c
}

func (s *PingSuite) Test_Ping() {

}

func Test_Ping(t *testing.T) {
	suite.Run(t, new(PingSuite))
}
