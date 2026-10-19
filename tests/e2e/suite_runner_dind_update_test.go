//go:build e2e_full

package e2e

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"
	"go.redsock.ru/toolbox"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	runnerDindUpdateSecondDindSuffix = "b"
	runnerDindUpdateSocketAddress    = "tcp://e2e-runner-dind-update-external:2375"
	runnerDindUpdateConcurrent       = int32(2)
)

func newUpdateRunnerDindRequest(runnerName, dindName string) *velez_api.UpdateRunnerConfig_Request {
	return &velez_api.UpdateRunnerConfig_Request{
		Name:     runnerName,
		DindName: toolbox.ToPtr(dindName),
	}
}

func newUpdateRunnerDindAndSocketRequest(runnerName, dindName, socketAddress string) *velez_api.UpdateRunnerConfig_Request {
	return &velez_api.UpdateRunnerConfig_Request{
		Name:                runnerName,
		DindName:            toolbox.ToPtr(dindName),
		DockerSocketAddress: toolbox.ToPtr(socketAddress),
	}
}

func newUpdateRunnerConcurrentRequest(runnerName string, concurrent int32) *velez_api.UpdateRunnerConfig_Request {
	return &velez_api.UpdateRunnerConfig_Request{
		Name:       runnerName,
		Concurrent: toolbox.ToPtr(concurrent),
	}
}

func newGetRunnerConfigRequest(runnerName string) *velez_api.GetRunnerConfig_Request {
	return &velez_api.GetRunnerConfig_Request{Name: runnerName}
}

func createSecondDind(t *testing.T, env *TestEnvironment, firstDindName string) string {
	t.Helper()

	secondName := firstDindName + runnerDindUpdateSecondDindSuffix

	registerDindCleanup(t, env, secondName)
	createDindAndAwait(t, env, newCreateDindRequest(secondName, toolbox.ToPtr(false)))

	return secondName
}

// createRunnerOnSocket creates a GitLab runner on a tcp:// docker_socket_address
// plus one idle DinD it can later be attached to.
func createRunnerOnSocket(t *testing.T, env *TestEnvironment) dindRunnerFixture {
	t.Helper()

	dockerClient := env.DockerClient()
	runnersClient := env.RunnersClient()

	dindName := dindE2eName(t)
	runnerName := dindE2eNamePrefixToRunner(dindName)
	runnerContainer := labels.GitlabRunnerNamePrefix + runnerName

	registerDindCleanup(t, env, dindName)

	t.Cleanup(func() {
		_, _ = runnersClient.DropRunner(context.Background(), newDropRunnerRequest(runnerContainer))

		removeOpts := container.RemoveOptions{Force: true, RemoveVolumes: true}
		_ = dockerClient.ContainerRemove(context.Background(), runnerContainer, removeOpts)
	})

	createDindAndAwait(t, env, newCreateDindRequest(dindName, toolbox.ToPtr(false)))

	stubName := gitlabStubContainerName(dindName)
	baseUrl := startGitlabStub(t, env, stubName)

	logContainerOnFailure(t, dockerClient, runnerContainer)

	req := newGitlabSocketCreateRunnerRequest(runnerName, baseUrl, runnerDindUpdateSocketAddress)

	created, err := runnersClient.CreateRunner(t.Context(), req)
	require.NoError(t, err)

	task := env.AwaitTask(t, created.GetEntityId(), created.GetAction())
	require.Equal(t, velez_api.TaskStatus_DONE, task.GetStatus(), "create runner task error: %s", task.GetError())

	return dindRunnerFixture{
		dindName:        dindName,
		runnerContainer: runnerContainer,
		stubName:        stubName,
	}
}

func dindE2eNamePrefixToRunner(dindName string) string {
	return "e2e-runner-" + dindName[len(dindE2eNamePrefix):]
}

func requireRunnerDind(t *testing.T, env *TestEnvironment, runnerContainer, wantDindName string) {
	t.Helper()

	config, err := env.RunnersClient().GetRunnerConfig(t.Context(), newGetRunnerConfigRequest(runnerContainer))
	require.NoError(t, err)
	require.Equal(t, wantDindName, config.GetDindName())
}

func runUpdateRunnerConfigAttachesDind(t *testing.T, env *TestEnvironment, _ Plane) {
	fixture := createRunnerOnSocket(t, env)

	requireRunnerDind(t, env, fixture.runnerContainer, "")

	resp, err := env.RunnersClient().UpdateRunnerConfig(
		t.Context(), newUpdateRunnerDindRequest(fixture.runnerContainer, fixture.dindName))
	require.NoError(t, err)
	require.True(t, resp.GetRequiresRedeploy())

	config, err := env.RunnersClient().GetRunnerConfig(t.Context(), newGetRunnerConfigRequest(fixture.runnerContainer))
	require.NoError(t, err)
	require.Equal(t, fixture.dindName, config.GetDindName())
	require.Empty(t, config.GetDockerSocketAddress())
}

func Test_Runner_UpdateConfig_AttachesDind(t *testing.T) {
	t.Parallel()

	runOnEveryPlane(t, runUpdateRunnerConfigAttachesDind)
}

func runUpdateRunnerConfigSwitchesDind(t *testing.T, env *TestEnvironment, _ Plane) {
	fixture := createRunnerOnDind(t, env)
	secondDindName := createSecondDind(t, env, fixture.dindName)

	resp, err := env.RunnersClient().UpdateRunnerConfig(
		t.Context(), newUpdateRunnerDindRequest(fixture.runnerContainer, secondDindName))
	require.NoError(t, err)
	require.True(t, resp.GetRequiresRedeploy())

	requireRunnerDind(t, env, fixture.runnerContainer, secondDindName)

	_, err = env.ServiceApiClient().RemoveService(t.Context(), newServiceRemoveRequest(secondDindName))
	require.Equal(t, codes.FailedPrecondition, status.Code(err))
	require.ErrorContains(t, err, "used by:")

	dropDindAndAwait(t, env, fixture.dindName)
	requireDindListed(t, env, secondDindName)
}

func Test_Runner_UpdateConfig_SwitchesDind(t *testing.T) {
	t.Parallel()

	runOnEveryPlane(t, runUpdateRunnerConfigSwitchesDind)
}

func runUpdateRunnerConfigDetachesDind(t *testing.T, env *TestEnvironment, _ Plane) {
	fixture := createRunnerOnDind(t, env)

	resp, err := env.RunnersClient().UpdateRunnerConfig(t.Context(), newUpdateRunnerDindRequest(fixture.runnerContainer, ""))
	require.NoError(t, err)
	require.True(t, resp.GetRequiresRedeploy())

	requireRunnerDind(t, env, fixture.runnerContainer, "")

	dropDindAndAwait(t, env, fixture.dindName)
}

func Test_Runner_UpdateConfig_DetachesDind(t *testing.T) {
	t.Parallel()

	runOnEveryPlane(t, runUpdateRunnerConfigDetachesDind)
}

func runUpdateRunnerConfigRejectsDindWithSocket(t *testing.T, env *TestEnvironment, _ Plane) {
	fixture := createRunnerOnDind(t, env)

	req := newUpdateRunnerDindAndSocketRequest(fixture.runnerContainer, fixture.dindName, runnerDindUpdateSocketAddress)

	_, err := env.RunnersClient().UpdateRunnerConfig(t.Context(), req)
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	requireRunnerDind(t, env, fixture.runnerContainer, fixture.dindName)
}

func Test_Runner_UpdateConfig_RejectsDindWithSocketAddress(t *testing.T) {
	t.Parallel()

	runOnEveryPlane(t, runUpdateRunnerConfigRejectsDindWithSocket)
}

func runUpdateRunnerConfigConcurrentKeepsDind(t *testing.T, env *TestEnvironment, _ Plane) {
	fixture := createRunnerOnDind(t, env)

	req := newUpdateRunnerConcurrentRequest(fixture.runnerContainer, runnerDindUpdateConcurrent)

	_, err := env.RunnersClient().UpdateRunnerConfig(t.Context(), req)
	require.NoError(t, err)

	requireRunnerDind(t, env, fixture.runnerContainer, fixture.dindName)

	_, err = env.ServiceApiClient().RemoveService(t.Context(), newServiceRemoveRequest(fixture.dindName))
	require.Equal(t, codes.FailedPrecondition, status.Code(err))
	require.ErrorContains(t, err, "used by:")
}

func Test_Runner_UpdateConfig_ConcurrentOnlyKeepsDind(t *testing.T) {
	t.Parallel()

	runOnEveryPlane(t, runUpdateRunnerConfigConcurrentKeepsDind)
}
