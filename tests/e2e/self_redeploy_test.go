//go:build e2e_full

package e2e

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

// SelfRedeploySuite covers the automatable slice of the SelfRedeploy guard: a
// redeploy that fails before the swap hand-off must leave no velez-swapper
// container behind.
//
// The transport-level not-in-container guard ("self redeploy is only available
// when velez runs inside a container") is deliberately NOT asserted here: since
// dcd5f593 the e2e harness runs Velez as a real container inside the DinD, so
// /.dockerenv is present, env.IsInContainer() is true (cached at app init by
// env.StartVolumes), and that guard can never fire in this fixture. The empty
// image happy path (the swapper kills and replaces the server under test) is
// verified manually per the card checklist. What remains automatable is the
// pre-hand-off failure: an unpullable image aborts the redeploy at the pull
// step - upstream of removeStaleSwapper/drainJobs/startSwapper - so a rejected
// call must leave the node untouched and no velez.swapper-labelled container
// may exist afterwards.
type SelfRedeploySuite struct {
	suite.Suite

	plane Plane
	ctx   context.Context
}

func (s *SelfRedeploySuite) SetupSuite() {
	s.ctx = context.Background()
}

// selfRedeployUnpullableImages are two distinct repository names that cannot
// be pulled (docker.io denies them). The second call proves a repeated
// rejection behaves identically to the first - no partial state, no swapper
// left behind by the failed attempt.
var selfRedeployUnpullableImages = []string{
	"velez-e2e-probe/nonexistent:0.0.1",
	"velez-e2e-probe/alsomissing:0.0.2",
}

func (s *SelfRedeploySuite) Test_UnpullableImage_RejectedWithoutSwapper() {
	t := s.T()

	env := s.plane.NewEnvironment(t)

	for _, image := range selfRedeployUnpullableImages {
		req := &velez_api.SelfRedeploy_Request{Image: image}

		resp, err := env.VelezApiClient().SelfRedeploy(s.ctx, req)
		require.Error(t, err)
		require.Nil(t, resp)
		require.Contains(t, err.Error(), "error pulling image for self redeploy")
	}

	requireNoSwapperContainer(t, env)
}

// requireNoSwapperContainer lists the DinD by the exact velez.swapper label
// (never by name substring - see docs/known_pitfalls.md) and fails on any
// hit, including exited containers.
func requireNoSwapperContainer(t *testing.T, env *TestEnvironment) {
	t.Helper()

	swapperFilter := filters.NewArgs(filters.Arg("label", labels.SwapperLabel))

	listOpts := container.ListOptions{All: true, Filters: swapperFilter}

	listed, err := env.DockerClient().ContainerList(t.Context(), listOpts)
	require.NoError(t, err)
	require.Empty(t, listed, "no %q-labelled container may exist after a rejected self redeploy", labels.SwapperLabel)
}

func Test_SelfRedeploy(t *testing.T) {
	RunPlaneSuite(t, Planes, func(plane Plane) suite.TestingSuite {
		return &SelfRedeploySuite{plane: plane}
	})
}
