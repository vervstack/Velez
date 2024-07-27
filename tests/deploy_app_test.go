//go:build integration

package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/godverv/Velez/pkg/velez_api"
)

const (
	appImage      = "godverv/hello_world:v0.0.8"
	postgresImage = "postgres:16"
)

type DeployAppSuite struct {
	suite.Suite
	ctx context.Context
}

func (d *DeployAppSuite) SetupSuite() {
	d.ctx = context.Background()
}

func (d *DeployAppSuite) Test_OK_DeployWithoutConfig() {
	serviceName := "simple_deploy"

	createReq := &velez_api.CreateSmerd_Request{
		Name:         serviceName,
		ImageName:    appImage,
		IgnoreConfig: true,
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

func (d *DeployAppSuite) Test_OK_DeployWithHealthCheck() {
	serviceName := "simple_deploy_with_health_check"

	createReq := &velez_api.CreateSmerd_Request{
		Name:      serviceName,
		ImageName: appImage,
		Healthcheck: &velez_api.Container_Healthcheck{
			IntervalSecond: 1,
			Retries:        3,
		},
		IgnoreConfig: true,
	}
	smerd, err := testEnv.callCreate(d.ctx, createReq)
	d.NoError(err)

	d.NotNil(smerd)
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

func (d *DeployAppSuite) Test_Fail_Deploy_NoConfig() {
	serviceName := "simple_deploy_without_config"

	createReq := &velez_api.CreateSmerd_Request{
		Name:      serviceName,
		ImageName: appImage,
	}
	smerd, err := testEnv.callCreate(d.ctx, createReq)

	grpcErr, ok := status.FromError(err)
	d.True(ok)
	d.Equal(codes.NotFound.String(), grpcErr.Code().String())

	d.Nil(smerd)
}

func (d *DeployAppSuite) Test_OK_DeployPostgres() {
	serviceName := "simple_deploy_postgres"

	timeout := uint32(5)
	command := "pg_isready -U postgres"

	createReq := &velez_api.CreateSmerd_Request{
		Name:      serviceName,
		ImageName: postgresImage,
		Healthcheck: &velez_api.Container_Healthcheck{
			Command:        &command,
			IntervalSecond: 2,
			TimeoutSecond:  &timeout,
			Retries:        3,
		},
		IgnoreConfig:  true,
		UseImagePorts: true,
	}
	smerd, err := testEnv.callCreate(d.ctx, createReq)
	d.NoError(err)

	d.NotNil(smerd)
	d.NotEmpty(smerd.Uuid)
	d.NotEmpty(smerd.CreatedAt)

	smerd.Uuid = ""
	smerd.CreatedAt = nil

	//expectedSmerd := &velez_api.Smerd{
	//	Name:      "/" + serviceName,
	//	ImageName: postgresImage,
	//	Status:    velez_api.Smerd_running,
	//	Labels:    testEnv.getExpectedLabels(),
	//}
	//d.Equal(smerd, expectedSmerd)
}

func (d *DeployAppSuite) TearDownSuite() {

}

func Test_DeployApp(t *testing.T) {
	suite.Run(t, new(DeployAppSuite))
}
