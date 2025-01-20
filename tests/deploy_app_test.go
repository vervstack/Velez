//go:build integration

package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/proto"

	"github.com/godverv/Velez/pkg/velez_api"
)

const (
	appImage      = "godverv/hello_world:v0.0.13"
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
	launchedSmerd, err := tEnv.callCreate(d.ctx, createReq)
	d.Require().NoError(err)

	d.Require().NotEmpty(launchedSmerd.Uuid)
	d.Require().NotEmpty(launchedSmerd.CreatedAt)

	launchedSmerd.Uuid = ""
	launchedSmerd.CreatedAt = nil

	expectedSmerd := &velez_api.Smerd{
		Name:      "/" + serviceName,
		ImageName: appImage,
		Status:    velez_api.Smerd_running,
		Labels:    tEnv.getExpectedLabels(),
	}

	if !proto.Equal(expectedSmerd, launchedSmerd) {
		d.Require().Equal(launchedSmerd, expectedSmerd)
	}
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
	launchedSmerd, err := tEnv.callCreate(d.ctx, createReq)
	d.Require().NoError(err)

	d.Require().NotNil(launchedSmerd)
	d.Require().NotEmpty(launchedSmerd.Uuid)
	d.Require().NotEmpty(launchedSmerd.CreatedAt)

	launchedSmerd.Uuid = ""
	launchedSmerd.CreatedAt = nil

	expectedSmerd := &velez_api.Smerd{
		Name:      "/" + serviceName,
		ImageName: appImage,
		Status:    velez_api.Smerd_running,
		Labels:    tEnv.getExpectedLabels(),
	}

	if !proto.Equal(expectedSmerd, launchedSmerd) {
		d.Require().Equal(launchedSmerd, expectedSmerd)
	}
}

func (d *DeployAppSuite) Test_OK_DeployWithDefaultConfig() {
	serviceName := "simple_deploy_with_default_config"

	createReq := &velez_api.CreateSmerd_Request{
		Name:      serviceName,
		ImageName: appImage,
	}
	smerd, err := tEnv.callCreate(d.ctx, createReq)
	d.Require().NoError(err)
	d.Require().NotNil(smerd)
}

func (d *DeployAppSuite) Test_OK_DeployPostgres() {
	serviceName := "simple_deploy_postgres"

	timeout := uint32(5)
	command := "pg_isready -U postgres"

	createReq := &velez_api.CreateSmerd_Request{
		Name:      serviceName,
		ImageName: postgresImage,
		Env: map[string]string{
			"POSTGRES_HOST_AUTH_METHOD": "trust",
		},
		Healthcheck: &velez_api.Container_Healthcheck{
			Command:        &command,
			IntervalSecond: 2,
			TimeoutSecond:  &timeout,
			Retries:        3,
		},
		IgnoreConfig:  true,
		UseImagePorts: true,
	}
	deployedSmerd, err := tEnv.callCreate(d.ctx, createReq)
	d.Require().NoError(err)

	d.Require().NotNil(deployedSmerd)
	d.Require().NotEmpty(deployedSmerd.Uuid)
	d.Require().NotEmpty(deployedSmerd.CreatedAt)

	deployedSmerd.Uuid = ""
	deployedSmerd.CreatedAt = nil

	expectedSmerd := &velez_api.Smerd{
		Name:      "/" + serviceName,
		ImageName: postgresImage,
		Status:    velez_api.Smerd_running,
		Labels:    tEnv.getExpectedLabels(),
		Ports: []*velez_api.Port{
			{
				ServicePortNumber: 0,
				Protocol:          0,
				ExposedTo:         nil,
			},
		},
	}

	d.Require().Equal(1, len(deployedSmerd.Ports))
	d.Require().Equal(uint32(5432), deployedSmerd.Ports[0].ServicePortNumber)
	d.Require().Equal(velez_api.Port_tcp, deployedSmerd.Ports[0].Protocol)
	d.Require().NotNil(*deployedSmerd.Ports[0].ExposedTo)
	d.Require().GreaterOrEqual(*deployedSmerd.Ports[0].ExposedTo, minPortToExposeTo)

	deployedSmerd.Ports = nil
	expectedSmerd.Ports = nil

	if !proto.Equal(expectedSmerd, deployedSmerd) {
		d.Require().Equal(expectedSmerd, deployedSmerd)
	}
}

func Test_DeployApp(t *testing.T) {
	suite.Run(t, new(DeployAppSuite))
}
