//go:build integration

package tests

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	appImage = "godverv/hello_world:v0.0.5ы"
)

type DeployAppSuite struct {
	suite.Suite
}

func (d *DeployAppSuite) SetupSuite() {

}

func (d *DeployAppSuite) Test_SimpleDeploy() {
	testEnv.grpcApi
}

func (d *DeployAppSuite) TearDownSuite() {

}

func Test_DeployApp(t *testing.T) {
	suite.Run(t, new(DeployAppSuite))
}
