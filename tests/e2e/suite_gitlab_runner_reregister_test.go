//go:build e2e_full

package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/gitlab_runner_config"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

func newReregisterRunnerRequest(name string) *velez_api.ReregisterRunner_Request {
	return &velez_api.ReregisterRunner_Request{Name: name}
}

func Test_GitlabRunner_Reregister_KeepsSingleRunnerEntry(t *testing.T) {
	t.Parallel()

	runOnEveryPlane(t, runGitlabRunnerReregisterKeepsSingleEntry)
}

func runGitlabRunnerReregisterKeepsSingleEntry(t *testing.T, env *TestEnvironment, _ Plane) {
	runnersClient := velez_api.NewRunnersAPIClient(env.grpcConn)
	fixture := createRunnerOnDind(t, env)

	requireGitlabRunnerDefaults(t, readRunnerConfig(t, env, fixture.runnerContainer), fixture.runnerContainer)

	for range 2 {
		reregistered, err := runnersClient.ReregisterRunner(t.Context(), newReregisterRunnerRequest(fixture.runnerContainer))
		require.NoError(t, err)

		task := awaitJobTask(t, env, reregistered.GetEntityId(), reregistered.GetAction())
		require.Equal(t, tasks_queries.VelezTaskStatusDONE, task.Status, "reregister task error: %s", task.Error.String)

		config := readRunnerConfig(t, env, fixture.runnerContainer)

		requireGitlabRunnerDefaults(t, config, fixture.runnerContainer)
	}
}

func readRunnerConfig(t *testing.T, env *TestEnvironment, runnerContainer string) []byte {
	t.Helper()

	dockerAPI := env.Custom.NodeClients.Docker().Client()

	config, err := dockerutils.ReadFromContainer(t.Context(), dockerAPI, runnerContainer, gitlabRunnerConfigPath)
	require.NoError(t, err)

	return config
}

func requireGitlabRunnerDefaults(t *testing.T, config []byte, runnerContainer string) {
	t.Helper()

	require.Equal(t, 1, gitlab_runner_config.CountRunners(config), "config.toml:\n%s", config)

	raw := string(config)
	require.Contains(t, raw, "privileged = true")
	require.Contains(t, raw, `pull_policy = ["if-not-present"]`)
	require.Contains(t, raw, `allowed_pull_policies = ["if-not-present", "always"]`)
	require.Contains(t, raw, `volumes = ["`+runnerContainer+`-cache:/cache"]`)
}
