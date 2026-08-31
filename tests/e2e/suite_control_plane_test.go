package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/storage/environments"
)

type ControlPlaneSuite struct {
	suite.Suite
}

// ListEnvironments returns rows now, not names, and the configured
// environments always come on top of the seeded default one - a Velez node is
// never environment-less.
func (s *ControlPlaneSuite) Test_ListEnvironments_WithLocalStateConfig() {
	t := s.T()

	configuredEnvs := []string{"dev", "staging", "prod"}
	env := NewEnvironment(t, WithEnvironments(configuredEnvs))

	req := &pb.ListEnvironments_Request{}
	resp, err := env.Custom.ControlPlaneApiImpl.ListEnvironments(t.Context(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	names := environmentNames(resp.GetEnvironments())

	expected := append([]string{environments.DefaultEnvironmentName}, configuredEnvs...)
	require.ElementsMatch(t, expected, names)
}

// Without any configured environment a node still has exactly one: the default
// one, carrying this node's (here unset) ContainerSuffix. That's what makes an
// environment-less request resolvable.
func (s *ControlPlaneSuite) Test_ListEnvironments_EmptyLocalStateConfig() {
	t := s.T()

	env := NewEnvironment(t)

	req := &pb.ListEnvironments_Request{}
	resp, err := env.Custom.ControlPlaneApiImpl.ListEnvironments(t.Context(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.GetEnvironments(), 1)
	require.Equal(t, environments.DefaultEnvironmentName, resp.GetEnvironments()[0].GetName())
	require.Empty(t, resp.GetEnvironments()[0].GetSuffix())
}

func environmentNames(envs []*pb.Environment) []string {
	names := make([]string, 0, len(envs))

	for _, e := range envs {
		names = append(names, e.GetName())
	}

	return names
}

func Test_ControlPlane(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(ControlPlaneSuite))
}
