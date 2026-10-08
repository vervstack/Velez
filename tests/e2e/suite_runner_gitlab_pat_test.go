//go:build e2e_full

package e2e

import (
	"archive/tar"
	"context"
	"io"
	"regexp"
	"strings"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/gitlab_runner_config"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

const (
	patLegacyRegistrationToken = "GR1348941abc"
	patRunnerAuthToken         = "glrt-already-a-runner-token"
	patConfigFileMode          = 0o600

	patStubLookupRequest = "GET /api/v4/projects/acme%2Fapp " + runnerDindAccessToken
	patStubCreateRequest = "POST /api/v4/user/runners " + runnerDindAccessToken
	patStubDeleteRequest = "DELETE /api/v4/runners "
	patStubResetRequest  = "POST /api/v4/runners/reset_authentication_token "
)

type gitlabPatFixture struct {
	runnerName      string
	runnerContainer string
	stubName        string
	baseUrl         string
	socketAddress   string
}

func startGitlabPatFixture(t *testing.T, env *TestEnvironment) gitlabPatFixture {
	t.Helper()

	dockerClient := env.Custom.NodeClients.Docker().Client()
	runnersClient := velez_api.NewRunnersAPIClient(env.grpcConn)

	suffix := dindE2eName(t)
	runnerName := strings.Replace(suffix, dindE2eNamePrefix, "e2e-runner-", 1)
	runnerContainer := labels.GitlabRunnerNamePrefix + runnerName
	stubName := gitlabStubContainerName(suffix)

	t.Cleanup(func() {
		_, _ = runnersClient.DropRunner(context.Background(), newDropRunnerRequest(runnerContainer))

		removeOpts := container.RemoveOptions{Force: true, RemoveVolumes: true}
		_ = dockerClient.ContainerRemove(context.Background(), runnerContainer, removeOpts)
		_ = dockerClient.VolumeRemove(context.Background(), runnerContainer+"-data", true)
		_ = dockerClient.VolumeRemove(context.Background(), runnerContainer+"-cache", true)
	})

	baseUrl := startGitlabStub(t, env, stubName)

	logContainerOnFailure(t, dockerClient, runnerContainer)

	return gitlabPatFixture{
		runnerName:      runnerName,
		runnerContainer: runnerContainer,
		stubName:        stubName,
		baseUrl:         baseUrl,
		socketAddress:   "tcp://" + suffix + "-external:2375",
	}
}

func newGitlabPatCreateRequest(fixture gitlabPatFixture, accessToken string) *velez_api.CreateRunner_Request {
	req := newGitlabSocketCreateRunnerRequest(fixture.runnerName, fixture.baseUrl, fixture.socketAddress)
	req.GetGitlab().AccessToken = accessToken

	return req
}

func createGitlabPatRunner(t *testing.T, env *TestEnvironment, fixture gitlabPatFixture) tasks_queries.VelezTask {
	t.Helper()

	runnersClient := velez_api.NewRunnersAPIClient(env.grpcConn)

	req := newGitlabPatCreateRequest(fixture, runnerDindAccessToken)

	created, err := runnersClient.CreateRunner(t.Context(), req)
	require.NoError(t, err)

	task := awaitJobTask(t, env, created.GetEntityId(), created.GetAction())
	require.Equal(t, tasks_queries.VelezTaskStatusDONE, task.Status, "create runner task error: %s", task.Error.String)

	return task
}

func dropGitlabPatRunner(t *testing.T, env *TestEnvironment, fixture gitlabPatFixture) {
	t.Helper()

	runnersClient := velez_api.NewRunnersAPIClient(env.grpcConn)

	dropped, err := runnersClient.DropRunner(t.Context(), newDropRunnerRequest(fixture.runnerContainer))
	require.NoError(t, err)

	task := awaitJobTask(t, env, dropped.GetEntityId(), dropped.GetAction())
	require.Equal(t, tasks_queries.VelezTaskStatusDONE, task.Status, "drop runner task error: %s", task.Error.String)
}

func reregisterGitlabPatRunner(t *testing.T, env *TestEnvironment, fixture gitlabPatFixture) {
	t.Helper()

	runnersClient := velez_api.NewRunnersAPIClient(env.grpcConn)

	reregistered, err := runnersClient.ReregisterRunner(t.Context(), newReregisterRunnerRequest(fixture.runnerContainer))
	require.NoError(t, err)

	task := awaitJobTask(t, env, reregistered.GetEntityId(), reregistered.GetAction())
	require.Equal(t, tasks_queries.VelezTaskStatusDONE, task.Status, "reregister task error: %s", task.Error.String)
}

func readConfigWithMode(t *testing.T, env *TestEnvironment, runnerContainer string) (content string, mode int64) {
	t.Helper()

	dockerClient := env.Custom.NodeClients.Docker().Client()

	reader, _, err := dockerClient.CopyFromContainer(t.Context(), runnerContainer, gitlabRunnerConfigPath)
	require.NoError(t, err)

	defer func() { _ = reader.Close() }()

	tarReader := tar.NewReader(reader)

	header, err := tarReader.Next()
	require.NoError(t, err)

	body, err := io.ReadAll(tarReader)
	require.NoError(t, err)

	return string(body), header.Mode & 0o777
}

func findRunnerRow(t *testing.T, env *TestEnvironment, runnerName string) *velez_api.Runner {
	t.Helper()

	runnersClient := velez_api.NewRunnersAPIClient(env.grpcConn)

	listed, err := runnersClient.ListRunners(t.Context(), &velez_api.ListRunners_Request{})
	require.NoError(t, err)

	for _, runner := range listed.GetRunners() {
		if strings.Contains(runner.GetName(), runnerName) {
			return runner
		}
	}

	return nil
}

func requireStubRequest(t *testing.T, env *TestEnvironment, stubName, prefix string) {
	t.Helper()

	requests := gitlabStubRequests(t, env, stubName)

	for _, request := range requests {
		if strings.HasPrefix(request, prefix) {
			return
		}
	}

	require.Failf(t, "gitlab stub request missing", "want prefix %q in %v", prefix, requests)
}

func countStubRequests(t *testing.T, env *TestEnvironment, stubName, prefix string) int {
	t.Helper()

	count := 0

	for _, request := range gitlabStubRequests(t, env, stubName) {
		if strings.HasPrefix(request, prefix) {
			count++
		}
	}

	return count
}

func requireSingleRunnerWithToken(t *testing.T, env *TestEnvironment, runnerContainer, token string) {
	t.Helper()

	content, mode := readConfigWithMode(t, env, runnerContainer)

	require.Equal(t, 1, gitlab_runner_config.CountRunners([]byte(content)), "config.toml:\n%s", content)
	require.Regexp(t, `(?m)^token = ['"]`+regexp.QuoteMeta(token)+`['"]$`, content)
	require.EqualValues(t, patConfigFileMode, mode)
}

func Test_GitlabRunner_Create_RejectsNonPersonalAccessToken(t *testing.T) {
	t.Parallel()

	runOnEveryPlane(t, runGitlabRunnerRejectsNonPersonalAccessToken)
}

func runGitlabRunnerRejectsNonPersonalAccessToken(t *testing.T, env *TestEnvironment, _ Plane) {
	runnersClient := velez_api.NewRunnersAPIClient(env.grpcConn)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	fixture := startGitlabPatFixture(t, env)

	for _, token := range []string{patLegacyRegistrationToken, patRunnerAuthToken, ""} {
		req := newGitlabPatCreateRequest(fixture, token)

		_, err := runnersClient.CreateRunner(t.Context(), req)
		require.Equal(t, codes.InvalidArgument, status.Code(err), "token %q: %v", token, err)

		if token != "" {
			require.Contains(t, status.Convert(err).Message(), "personal access token")
		}
	}

	require.Empty(t, gitlabStubRequests(t, env, fixture.stubName), "nothing may reach gitlab for a rejected token")
	require.Nil(t, findRunnerRow(t, env, fixture.runnerName))

	listed, err := runnersClient.ListRunners(t.Context(), &velez_api.ListRunners_Request{})
	require.NoError(t, err)

	for _, provisioning := range listed.GetProvisioning() {
		require.NotContains(t, provisioning.GetEntityId(), fixture.runnerName, "a rejected create must enqueue nothing")
	}

	_, err = dockerClient.ContainerInspect(t.Context(), fixture.runnerContainer)
	require.Error(t, err, "a rejected create must deploy nothing")
}

func Test_GitlabRunner_Create_SeedsConfigAndKeepsPatOutOfTaskContext(t *testing.T) {
	t.Parallel()

	runOnEveryPlane(t, runGitlabRunnerCreateSeedsConfig)
}

func runGitlabRunnerCreateSeedsConfig(t *testing.T, env *TestEnvironment, _ Plane) {
	fixture := startGitlabPatFixture(t, env)

	task := createGitlabPatRunner(t, env, fixture)

	require.NotContains(t, string(task.Context.RawMessage), runnerDindAccessToken, "the PAT must not be persisted")

	requireStubRequest(t, env, fixture.stubName, patStubLookupRequest)
	requireStubRequest(t, env, fixture.stubName, patStubCreateRequest)

	require.NotNil(t, findRunnerRow(t, env, fixture.runnerName), "runner row must be listed")

	dataStorage := env.Custom.Services.StorageContainer()

	svc, err := dataStorage.Services().GetByName(t.Context(), fixture.runnerContainer)
	require.NoError(t, err)

	runner, err := dataStorage.Runners().GetRunnerByServiceID(t.Context(), svc.ID)
	require.NoError(t, err)
	require.EqualValues(t, gitlabStubRunnerId, runner.GitlabRunnerId)

	requireSingleRunnerWithToken(t, env, fixture.runnerContainer, gitlabStubRunnerToken)
}

func Test_GitlabRunner_Recreate_AfterDropWaitsForRealDeploy(t *testing.T) {
	t.Parallel()

	runOnEveryPlane(t, runGitlabRunnerRecreateAfterDrop)
}

func runGitlabRunnerRecreateAfterDrop(t *testing.T, env *TestEnvironment, _ Plane) {
	dockerClient := env.Custom.NodeClients.Docker().Client()

	fixture := startGitlabPatFixture(t, env)

	createGitlabPatRunner(t, env, fixture)

	first, err := dockerClient.ContainerInspect(t.Context(), fixture.runnerContainer)
	require.NoError(t, err)

	dropGitlabPatRunner(t, env, fixture)

	_, err = dockerClient.ContainerInspect(t.Context(), fixture.runnerContainer)
	require.Error(t, err, "dropped runner container must be gone")

	createGitlabPatRunner(t, env, fixture)

	second, err := dockerClient.ContainerInspect(t.Context(), fixture.runnerContainer)
	require.NoError(t, err, "create task finished before the runner container existed")
	require.NotEqual(t, first.ID, second.ID)
	require.True(t, second.State.Running, "create task finished before the runner container was running")

	require.Equal(t, 2, countStubRequests(t, env, fixture.stubName, patStubCreateRequest))
	requireSingleRunnerWithToken(t, env, fixture.runnerContainer, gitlabStubRunnerToken)
}

func Test_GitlabRunner_Drop_DeletesRunnerOnGitlabAndRemovesRow(t *testing.T) {
	t.Parallel()

	runOnEveryPlane(t, runGitlabRunnerDropDeletesOnGitlab)
}

func runGitlabRunnerDropDeletesOnGitlab(t *testing.T, env *TestEnvironment, _ Plane) {
	dockerClient := env.Custom.NodeClients.Docker().Client()

	fixture := startGitlabPatFixture(t, env)

	createGitlabPatRunner(t, env, fixture)
	require.Zero(t, countStubRequests(t, env, fixture.stubName, patStubDeleteRequest))

	dropGitlabPatRunner(t, env, fixture)

	requireStubRequest(t, env, fixture.stubName, patStubDeleteRequest)
	require.Nil(t, findRunnerRow(t, env, fixture.runnerName), "dropped runner row must be gone")

	_, err := dockerClient.ContainerInspect(t.Context(), fixture.runnerContainer)
	require.Error(t, err, "dropped runner container must be gone")

	dataStorage := env.Custom.Services.StorageContainer()

	_, err = dataStorage.Services().GetByName(t.Context(), fixture.runnerContainer)
	require.Error(t, err, "dropped runner service must be gone")
}

func Test_GitlabRunner_Reregister_RotatesTokenInConfig(t *testing.T) {
	t.Parallel()

	runOnEveryPlane(t, runGitlabRunnerReregisterRotatesToken)
}

func runGitlabRunnerReregisterRotatesToken(t *testing.T, env *TestEnvironment, _ Plane) {
	fixture := startGitlabPatFixture(t, env)

	createGitlabPatRunner(t, env, fixture)
	requireSingleRunnerWithToken(t, env, fixture.runnerContainer, gitlabStubRunnerToken)

	reregisterGitlabPatRunner(t, env, fixture)

	requireStubRequest(t, env, fixture.stubName, patStubResetRequest)
	requireSingleRunnerWithToken(t, env, fixture.runnerContainer, gitlabStubRotatedToken)

	content, _ := readConfigWithMode(t, env, fixture.runnerContainer)
	require.NotContains(t, content, gitlabStubRunnerToken)
}
