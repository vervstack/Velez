package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

type ControlPlaneSuite struct {
	suite.Suite
}

func (s *ControlPlaneSuite) Test_ListEnvironments_WithLocalStateConfig() {
	t := s.T()
	t.Parallel()

	expectedEnvs := []string{"dev", "staging", "prod"}
	env := NewEnvironment(t, WithEnvironments(expectedEnvs))

	req := &pb.ListEnvironments_Request{}
	resp, err := env.Custom.ControlPlaneApiImpl.ListEnvironments(t.Context(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.ElementsMatch(t, expectedEnvs, resp.Environments)
}

func (s *ControlPlaneSuite) Test_ListEnvironments_EmptyLocalStateConfig() {
	t := s.T()
	t.Parallel()

	env := NewEnvironment(t)

	req := &pb.ListEnvironments_Request{}
	resp, err := env.Custom.ControlPlaneApiImpl.ListEnvironments(t.Context(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Empty(t, resp.Environments)
}

func Test_ControlPlane(t *testing.T) {
	suite.Run(t, new(ControlPlaneSuite))
}
