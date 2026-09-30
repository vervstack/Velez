//go:build e2e_full

package e2e

import (
	"io"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	runnerRegTarget       = "acme/app"
	runnerRegEnvToken     = "abc"
	runnerRegAccessToken  = "e2e-access-token"
	runnerRegListLimit    = 100
	runnerRegClusterSufix = "-runner"
	runnerRegImageName    = "ghcr.io/actions/actions-runner:e2e"
)

var runnerRegForeignEnv = []string{
	"RUNNER_TOKEN=" + runnerRegEnvToken,
	"RUNNER_REPO_URL=https://github.com/" + runnerRegTarget,
}

func runnerRegNames(prefix string) (containerName, serviceName string) {
	return "e2e_regrunner_" + prefix + "_c", "e2e_regrunner_" + prefix + "_svc"
}

func newRunnerRegisterRequest(
	containerId, serviceName string, accessToken *string,
) *velez_api.RegisterContainer_Request {
	req := newRegisterContainerRequest(containerId, serviceName)

	pattern := &velez_api.RegisterContainer_Request_RunnerPattern{
		Provider:    velez_api.RunnerProvider_GITHUB,
		Scope:       velez_api.RunnerScope_REPO,
		Target:      runnerRegTarget,
		Labels:      []string{"e2e", "arm"},
		AccessToken: accessToken,
	}

	req.Pattern = &velez_api.RegisterContainer_Request_Runner{Runner: pattern}

	return req
}

func requireNoPendingRunnerSecrets(t *testing.T, env *TestEnvironment, entityId string) {
	t.Helper()

	refs := []domain.SecretRef{
		jobs.RegisteredRunnerPendingAccessTokenRef(entityId),
		jobs.RegisteredRunnerPendingRegistrationTokenRef(entityId),
	}

	for _, ref := range refs {
		_, err := env.Custom.Services.Secrets().Get(t.Context(), ref)
		require.True(t, rerrors.Is(err, user_errors.ErrSecretNotFound), "pending secret must be gone, got %v", err)
	}
}

func requireRunnerListed(t *testing.T, env *TestEnvironment, serviceName string) *velez_api.Runner {
	t.Helper()

	req := &velez_api.ListRunners_Request{Paging: &velez_api.Paging{Limit: runnerRegListLimit}}

	resp, err := env.Custom.RunnersApiImpl.ListRunners(t.Context(), req)
	require.NoError(t, err)

	for _, runner := range resp.GetRunners() {
		if runner.GetName() == serviceName {
			return runner
		}
	}

	require.Failf(t, "runner not listed", "runner %q missing from ListRunners", serviceName)

	return nil
}

func requireRunnerRegistrationToken(t *testing.T, env *TestEnvironment, serviceName string) {
	t.Helper()

	req := &velez_api.GetRunnerCredentials_Request{Name: serviceName}

	creds, err := env.Custom.RunnersApiImpl.GetRunnerCredentials(t.Context(), req)
	require.NoError(t, err)
	require.Equal(t, runnerRegEnvToken, creds.GetToken())
	require.Equal(t, runnerRegTarget, creds.GetTarget())
	require.Equal(t, velez_api.RunnerProvider_GITHUB, creds.GetProvider())
}

func Test_RegisterContainer_Runner_SingleMode(t *testing.T) {
	t.Parallel()

	containerName, serviceName := runnerRegNames("single")

	env := Planes[0].NewEnvironment(t)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	foreign := startForeignContainerWithEnv(t, env, containerName, false, runnerRegForeignEnv)

	req := newRunnerRegisterRequest(foreign.id, serviceName, toolbox.ToPtr(runnerRegAccessToken))
	resp := registerContainer(t, env, req)

	task := awaitRegisterTask(t, env, resp)
	require.Equal(t, tasks_queries.VelezTaskStatusDONE, task.Status, "register task error: %s", task.Error.String)
	requireNoPasswordInPayload(t, task, runnerRegAccessToken, runnerRegEnvToken)

	inspected, err := dockerClient.ContainerInspect(t.Context(), containerName)
	require.NoError(t, err)
	require.True(t, inspected.State.Running)
	require.NotEqual(t, foreign.id, inspected.ID, "single mode recreates the container")

	containerLabels := inspected.Config.Labels
	require.Equal(t, serviceName, containerLabels[labels.VervServiceLabel])
	require.Equal(t, labelValueTrue, containerLabels[labels.RunnerInstanceLabel])
	require.Equal(t, velez_api.RunnerProvider_GITHUB.String(), containerLabels[labels.RunnerProviderLabel])
	require.Equal(t, velez_api.RunnerScope_REPO.String(), containerLabels[labels.RunnerScopeLabel])
	require.Equal(t, runnerRegTarget, containerLabels[labels.RunnerTargetLabel])
	require.Equal(t, "e2e,arm", containerLabels[labels.RunnerLabelsLabel])

	require.Contains(t, inspected.Config.Env, "VELEZ_RUNNER_REGISTRATION_TOKEN="+runnerRegEnvToken)

	runner := requireRunnerListed(t, env, serviceName)
	require.Equal(t, runnerRegTarget, runner.GetTarget())

	requireRunnerRegistrationToken(t, env, serviceName)
	requireNoPendingRunnerSecrets(t, env, resp.GetEntityId())
}

func Test_RegisterContainer_Runner_ClusterMode(t *testing.T) {
	containerName, serviceName := runnerRegNames("cluster")

	env, _ := enableStatefullPgUnderDind(t, Planes[1], registerClusterSuffix+runnerRegClusterSufix)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	foreign := startForeignContainerWithEnv(t, env, containerName, true, runnerRegForeignEnv)

	req := newRunnerRegisterRequest(foreign.id, serviceName, toolbox.ToPtr(runnerRegAccessToken))
	resp := registerContainer(t, env, req)

	task := awaitRegisterTask(t, env, resp)
	require.Equal(t, tasks_queries.VelezTaskStatusDONE, task.Status, "register task error: %s", task.Error.String)
	requireNoPasswordInPayload(t, task, runnerRegAccessToken, runnerRegEnvToken)

	requireContainerUntouched(t, dockerClient, foreign)

	runner := requireRunnerListed(t, env, serviceName)
	require.Equal(t, runnerRegTarget, runner.GetTarget())
	require.Equal(t, velez_api.RunnerProvider_GITHUB, runner.GetProvider())
	require.Equal(t, []string{"e2e", "arm"}, runner.GetLabels())

	requireRunnerRegistrationToken(t, env, serviceName)

	accessToken, err := env.Custom.Services.Secrets().Get(t.Context(), domain.RunnerAccessTokenSecretRef(serviceName))
	require.NoError(t, err)
	require.Equal(t, runnerRegAccessToken, accessToken)

	requireDeploymentRunning(t, env, serviceName)
	requireNoPendingRunnerSecrets(t, env, resp.GetEntityId())
}

func Test_RegisterContainer_Runner_SuggestsDefaultsFromContainer(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t)

	containerName, _ := runnerRegNames("suggest")

	foreign := startForeignRunnerImageContainer(t, env, containerName)

	req := &velez_api.GetContainer_Request{Id: foreign.id}

	got, err := env.Custom.ApiGrpcImpl.GetContainer(t.Context(), req)
	require.NoError(t, err)
	require.Equal(t, velez_api.ServicePattern_SERVICE_PATTERN_GITHUB_RUNNER, got.GetSuggestedPattern())

	defaults := got.GetSuggestedRunnerDefaults()
	require.Equal(t, velez_api.RunnerProvider_GITHUB.String(), defaults.GetProvider())
	require.Equal(t, velez_api.RunnerScope_REPO.String(), defaults.GetScope())
	require.Equal(t, runnerRegTarget, defaults.GetTarget())
	require.True(t, defaults.GetIsRegistrationTokenFound())

	plain := startForeignContainer(t, env, containerName+"_plain", false)

	plainGot, err := env.Custom.ApiGrpcImpl.GetContainer(t.Context(), &velez_api.GetContainer_Request{Id: plain.id})
	require.NoError(t, err)
	require.Nil(t, plainGot.GetSuggestedRunnerDefaults())
}

// startForeignRunnerImageContainer runs nginx under the actions-runner image
// name, which is all the suggested-pattern detection looks at.
func startForeignRunnerImageContainer(t *testing.T, env *TestEnvironment, name string) foreignContainer {
	t.Helper()

	dockerClient := env.Custom.NodeClients.Docker().Client()

	pullReader, err := dockerClient.ImagePull(t.Context(), NginxAlpineImage, image.PullOptions{})
	require.NoError(t, err)

	_, err = io.Copy(io.Discard, pullReader)
	require.NoError(t, err)

	err = pullReader.Close()
	require.NoError(t, err)

	err = dockerClient.ImageTag(t.Context(), NginxAlpineImage, runnerRegImageName)
	require.NoError(t, err)

	cfg := &container.Config{
		Image:  runnerRegImageName,
		Env:    runnerRegForeignEnv,
		Labels: map[string]string{testCaseNameLabel: t.Name()},
	}

	t.Cleanup(func() {
		removeForeignContainer(dockerClient, name, "")
	})

	created, err := dockerClient.ContainerCreate(t.Context(), cfg, nil, nil, nil, name)
	require.NoError(t, err)

	err = dockerClient.ContainerStart(t.Context(), created.ID, container.StartOptions{})
	require.NoError(t, err)

	return foreignContainer{id: created.ID, name: name}
}

func Test_RegisterContainer_Runner_RejectsBadRequests(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t)
	velezClient := velez_api.NewVelezAPIClient(env.grpcConn)

	t.Run("target is required", func(t *testing.T) {
		t.Parallel()

		req := newRunnerRegisterRequest("some-id", "e2e_regrunner_bad", nil)

		req.GetRunner().Target = ""

		_, err := velezClient.RegisterContainer(t.Context(), req)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("provider is required", func(t *testing.T) {
		t.Parallel()

		req := newRunnerRegisterRequest("some-id", "e2e_regrunner_bad", nil)

		req.GetRunner().Provider = velez_api.RunnerProvider_RUNNER_PROVIDER_UNSPECIFIED

		_, err := velezClient.RegisterContainer(t.Context(), req)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	})
}
