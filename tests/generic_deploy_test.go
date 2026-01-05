package tests

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type GenericDeploySuite struct {
	suite.Suite
}

func (s *GenericDeploySuite) SetupSuite() {
	t := s.T()
	env := NewEnvironment(t)
	_ = env
	println(123)
}

func (s *GenericDeploySuite) Test_DeployPostgres() {
	t := s.T()
	t.Skip("skipping for now")
}

func Test_GenericDeploy(t *testing.T) {
	suite.Run(t, new(GenericDeploySuite))
}
