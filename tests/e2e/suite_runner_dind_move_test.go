//go:build e2e_full

package e2e

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/jobs"
)

const (
	runnerMoveBuildkitWindow = 4 * time.Minute
	runnerMoveListWindow     = 15 * time.Second
)

// buildkitMoveFixture is a runner with BuildKit enabled on its DinD, plus a
// second DinD to move it to.
type buildkitMoveFixture struct {
	dindRunnerFixture
	secondDindName string
}

func newBuildkitMoveFixture(t *testing.T, env *TestEnvironment) buildkitMoveFixture {
	t.Helper()

	fixture := createRunnerOnDind(t, env)
	secondDindName := createSecondDind(t, env, fixture.dindName)

	logDindOnFailure(t, env, fixture.dindName, fixture.runnerContainer)
	logDindOnFailure(t, env, secondDindName, fixture.runnerContainer)
	preloadBuildkitImage(t, env, fixture.dindName)
	preloadBuildkitImage(t, env, secondDindName)

	setRunnerBuildkitAndAwait(t, env, fixture.runnerContainer, true)
	requireBuildkitInsideDind(t, env, fixture.dindName, fixture.runnerContainer)

	return buildkitMoveFixture{dindRunnerFixture: fixture, secondDindName: secondDindName}
}

func requireRunnerBuildkitFlag(t *testing.T, env *TestEnvironment, runnerContainer string, want bool) {
	t.Helper()

	require.Eventually(t, func() bool {
		listResp, err := env.RunnersClient().ListRunners(t.Context(), &velez_api.ListRunners_Request{})
		if err != nil {
			return false
		}

		for _, runner := range listResp.GetRunners() {
			if runner.GetName() == runnerContainer {
				return runner.GetIsBuildkitEnabled() == want
			}
		}

		return false
	}, runnerMoveListWindow, time.Second, "runner %q buildkit flag never became %v", runnerContainer, want)
}

func requireBuildkitdAbsentFromDind(t *testing.T, env *TestEnvironment, dindName, runnerContainer string) {
	t.Helper()

	buildkitName := domain.RunnerBuildkitServiceName(runnerContainer)
	volumeName := domain.RunnerBuildkitStateVolumeName(runnerContainer)

	containers := innerDockerLines(t, env, dindName, "ps", "-a", "--format", "{{.Names}}")
	require.NotContains(t, containers, buildkitName)

	networks := innerDockerLines(t, env, dindName, "network", "ls", "--format", "{{.Name}}")
	require.NotContains(t, networks, buildkitName)

	volumes := innerDockerLines(t, env, dindName, "volume", "ls", "--format", "{{.Name}}")
	require.NotContains(t, volumes, volumeName)
}

func requireBuildkitdRunningInDind(t *testing.T, env *TestEnvironment, dindName, runnerContainer string) {
	t.Helper()

	buildkitName := domain.RunnerBuildkitServiceName(runnerContainer)

	require.Eventually(t, func() bool {
		output, exitCode := execInDind(t, env, dindName, "ps", "--format", "{{.Names}}")

		return exitCode == 0 && slices.Contains(strings.Split(output, "\n"), buildkitName)
	}, runnerMoveBuildkitWindow, buildkitProbeTick, "buildkitd never ran in dind %q", dindName)
}

func requireRunnerDindLabel(t *testing.T, inspected container.InspectResponse, want string) {
	t.Helper()

	got, hasLabel := inspected.Config.Labels[labels.RunnerDindLabel]
	if want == "" {
		require.False(t, hasLabel, "runner must carry no dind label, got %q", got)

		return
	}

	require.Equal(t, want, got)
}

func requireRunnerLeftDind(t *testing.T, inspected container.InspectResponse, dindName string) {
	t.Helper()

	require.NotContains(t, inspected.NetworkSettings.Networks, domain.DindNetworkName(dindName))
	require.NotContains(t, inspected.Config.Env, "DOCKER_HOST="+domain.DindAddress(dindName))
}

func updateRunnerDind(t *testing.T, env *TestEnvironment, runnerContainer, dindName string) {
	t.Helper()

	resp, err := env.RunnersClient().UpdateRunnerConfig(t.Context(), newUpdateRunnerDindRequest(runnerContainer, dindName))
	require.NoError(t, err)
	require.True(t, resp.GetRequiresRedeploy())
}

func Test_RunnerBuildkit_SwitchDind_MovesBuildkitAlong(t *testing.T) {
	t.Parallel()

	runOnEveryPlane(t, runRunnerBuildkitSwitchDind)
}

func runRunnerBuildkitSwitchDind(t *testing.T, env *TestEnvironment, _ Plane) {
	fixture := newBuildkitMoveFixture(t, env)

	updateRunnerDind(t, env, fixture.runnerContainer, fixture.secondDindName)

	requireBuildkitdAbsentFromDind(t, env, fixture.dindName, fixture.runnerContainer)
	requireRunnerBuildkitFlag(t, env, fixture.runnerContainer, true)
	requireRunnerNetworkMode(t, env, fixture.runnerContainer, domain.RunnerBuildkitServiceName(fixture.runnerContainer))

	redeployed := redeployRunnerAndAwaitNewContainer(t, env, fixture.runnerContainer, fixture.inspected.ID)

	requireRunnerOnDind(t, redeployed, fixture.secondDindName)
	requireRunnerDindLabel(t, redeployed, fixture.secondDindName)
	requireRunnerLeftDind(t, redeployed, fixture.dindName)

	requireBuildkitdRunningInDind(t, env, fixture.secondDindName, fixture.runnerContainer)

	task := env.AwaitTask(t, fixture.runnerContainer, jobs.SetRunnerBuildkitAction)
	require.Equal(t, velez_api.TaskStatus_DONE, task.GetStatus(), "set runner buildkit task error: %s", task.GetError())

	requireBuildkitInsideDind(t, env, fixture.secondDindName, fixture.runnerContainer)
	requireBuildkitdAbsentFromDind(t, env, fixture.dindName, fixture.runnerContainer)
	requireRunnerBuildkitFlag(t, env, fixture.runnerContainer, true)
}

func Test_RunnerBuildkit_DetachDind_TurnsBuildkitOff(t *testing.T) {
	t.Parallel()

	runOnEveryPlane(t, runRunnerBuildkitDetachDind)
}

func runRunnerBuildkitDetachDind(t *testing.T, env *TestEnvironment, _ Plane) {
	fixture := newBuildkitMoveFixture(t, env)

	updateRunnerDind(t, env, fixture.runnerContainer, "")

	requireBuildkitGoneFromDind(t, env, fixture.dindName, fixture.runnerContainer)
	requireRunnerBuildkitFlag(t, env, fixture.runnerContainer, false)
	requireRunnerDind(t, env, fixture.runnerContainer, "")

	redeployed := redeployRunnerAndAwaitNewContainer(t, env, fixture.runnerContainer, fixture.inspected.ID)

	requireRunnerDindLabel(t, redeployed, "")
	requireRunnerLeftDind(t, redeployed, fixture.dindName)
	requireBuildkitGoneFromDind(t, env, fixture.dindName, fixture.runnerContainer)
}

func Test_Runner_UpdateConfig_RedeployRewritesDindLabel(t *testing.T) {
	t.Parallel()

	runOnEveryPlane(t, runUpdateRunnerConfigRedeployRewritesDindLabel)
}

func runUpdateRunnerConfigRedeployRewritesDindLabel(t *testing.T, env *TestEnvironment, _ Plane) {
	fixture := createRunnerOnDind(t, env)
	secondDindName := createSecondDind(t, env, fixture.dindName)

	updateRunnerDind(t, env, fixture.runnerContainer, secondDindName)

	switched := redeployRunnerAndAwaitNewContainer(t, env, fixture.runnerContainer, fixture.inspected.ID)

	requireRunnerOnDind(t, switched, secondDindName)
	requireRunnerDindLabel(t, switched, secondDindName)
	requireRunnerLeftDind(t, switched, fixture.dindName)

	updateRunnerDind(t, env, fixture.runnerContainer, "")

	detached := redeployRunnerAndAwaitNewContainer(t, env, fixture.runnerContainer, switched.ID)

	requireRunnerDindLabel(t, detached, "")
	requireRunnerLeftDind(t, detached, secondDindName)
	requireNoDockerSocketMount(t, detached)
}

func Test_Runner_UpdateConfig_AttachRedeployRewritesDindLabel(t *testing.T) {
	t.Parallel()

	runOnEveryPlane(t, runUpdateRunnerConfigAttachRedeployRewritesDindLabel)
}

func runUpdateRunnerConfigAttachRedeployRewritesDindLabel(t *testing.T, env *TestEnvironment, _ Plane) {
	fixture := createRunnerOnSocket(t, env)

	before, err := env.DockerClient().ContainerInspect(t.Context(), fixture.runnerContainer)
	require.NoError(t, err)
	requireRunnerDindLabel(t, before, "")

	updateRunnerDind(t, env, fixture.runnerContainer, fixture.dindName)

	attached := redeployRunnerAndAwaitNewContainer(t, env, fixture.runnerContainer, before.ID)

	requireRunnerOnDind(t, attached, fixture.dindName)
	requireRunnerDindLabel(t, attached, fixture.dindName)
}
