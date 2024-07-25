//go:build integration

package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/godverv/Velez/pkg/velez_api"
)

const (
	appImage = "godverv/hello_world:v0.0.6"
)

type DeployAppSuite struct {
	suite.Suite
	ctx context.Context
}

func (d *DeployAppSuite) SetupSuite() {
	d.ctx = context.Background()
}

func (d *DeployAppSuite) Test_SimpleDeploy() {
	serviceName := "simple_deploy"

	createReq := &velez_api.CreateSmerd_Request{
		Name:      serviceName,
		ImageName: appImage,
	}
	smerd, err := testEnv.callCreate(d.ctx, createReq)
	d.NoError(err)

	d.NotEmpty(smerd.Uuid)
	d.NotEmpty(smerd.CreatedAt)

	smerd.Uuid = ""
	smerd.CreatedAt = nil

	expectedSmerd := &velez_api.Smerd{
		Name:      "/" + serviceName,
		ImageName: appImage,
		Status:    velez_api.Smerd_running,
		Labels:    testEnv.getExpectedLabels(),
	}
	d.Equal(smerd, expectedSmerd)
}

func (d *DeployAppSuite) Test_DeployWithHealthCheck() {
	serviceName := "simple_deploy_with_health_check"

	createReq := &velez_api.CreateSmerd_Request{
		Name:      serviceName,
		ImageName: appImage,
		Healthcheck: &velez_api.Container_Healthcheck{
			IntervalSecond: 5,
			Retries:        3,
		},
	}
	smerd, err := testEnv.callCreate(d.ctx, createReq)
	d.NoError(err)

	d.NotEmpty(smerd.Uuid)
	d.NotEmpty(smerd.CreatedAt)

	smerd.Uuid = ""
	smerd.CreatedAt = nil

	expectedSmerd := &velez_api.Smerd{
		Name:      "/" + serviceName,
		ImageName: appImage,
		Status:    velez_api.Smerd_running,
		Labels:    testEnv.getExpectedLabels(),
	}
	d.Equal(smerd, expectedSmerd)
}

func (d *DeployAppSuite) TearDownSuite() {

}

func Test_DeployApp(t *testing.T) {
	suite.Run(t, new(DeployAppSuite))
}
