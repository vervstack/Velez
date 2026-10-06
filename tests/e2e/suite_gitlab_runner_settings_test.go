//go:build e2e_full

package e2e

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/gitlab_runner_config"
	"go.vervstack.ru/Velez/internal/service/service_manager/runneraas/providers/gitlab"
)

const (
	gitlabRunnerGeneratedConfig = `concurrent = 1
check_interval = 0
shutdown_timeout = 0

[session_server]
  session_timeout = 1800

[[runners]]
  name = "x"
  url = "https://gitlab.com"
  token = "glrt-e2e-token"
  executor = "docker"
  [runners.docker]
    tls_verify = false
    image = "alpine:latest"
    privileged = true
`

	gitlabRunnerFreshEntry = `
[[runners]]
  name = "y"
  url = "https://gitlab.com"
  token = "glrt-e2e-token-2"
  executor = "docker"
  [runners.docker]
    tls_verify = false
    image = "alpine:latest"
`

	gitlabRunnerConcurrentOnly = "concurrent = 1\n"

	pullAlways       = "always"
	pullIfNotPresent = "if-not-present"
)

func Test_GitlabRunner_ApplySettings_WritesPullPolicyAndGlobalKeys(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	runtime, name := startGitlabRunnerContainer(t, gitlabRunnerGeneratedConfig)
	settings := newFullRunnerSettings()

	err := gitlab.New().ApplySettings(ctx, runtime, name, settings)
	require.NoError(t, err)

	got, err := gitlab.New().ReadSettings(ctx, runtime, name)
	require.NoError(t, err)
	require.Equal(t, settings, got)

	raw := readGitlabRunnerConfig(t, runtime, name)
	require.Contains(t, raw, "concurrent = 1")
	require.Contains(t, raw, `name = "x"`)
	require.Contains(t, raw, "privileged = true")

	requireGitlabRunnerListOk(t, runtime, name)
}

func Test_GitlabRunner_ApplySettings_ClearsKeys(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	runtime, name := startGitlabRunnerContainer(t, gitlabRunnerGeneratedConfig)
	provider := gitlab.New()

	err := provider.ApplySettings(ctx, runtime, name, newFullRunnerSettings())
	require.NoError(t, err)

	err = provider.ApplySettings(ctx, runtime, name, domain.GitlabRunnerSettings{})
	require.NoError(t, err)

	got, err := provider.ReadSettings(ctx, runtime, name)
	require.NoError(t, err)
	require.Equal(t, domain.GitlabRunnerSettings{}, got)

	raw := readGitlabRunnerConfig(t, runtime, name)
	require.NotContains(t, raw, "pull_policy")
	require.NotContains(t, raw, "allowed_pull_policies")
	require.NotContains(t, raw, "log_level")

	requireGitlabRunnerListOk(t, runtime, name)
}

func Test_GitlabRunner_ApplySettings_SurvivesReregister(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	runtime, name := startGitlabRunnerContainer(t, gitlabRunnerGeneratedConfig)
	provider := gitlab.New()

	err := provider.ApplySettings(ctx, runtime, name, newFullRunnerSettings())
	require.NoError(t, err)

	snapshot, err := provider.ReadSettings(ctx, runtime, name)
	require.NoError(t, err)

	unregistered := []byte(gitlabRunnerConcurrentOnly)

	err = runtime.CopyToContainer(ctx, name, gitlabRunnerConfigPath, unregistered, 0o600)
	require.NoError(t, err)

	reregistered := []byte(gitlabRunnerConcurrentOnly + gitlabRunnerFreshEntry)

	err = runtime.CopyToContainer(ctx, name, gitlabRunnerConfigPath, reregistered, 0o600)
	require.NoError(t, err)

	err = provider.ApplySettings(ctx, runtime, name, snapshot)
	require.NoError(t, err)

	got, err := provider.ReadSettings(ctx, runtime, name)
	require.NoError(t, err)
	require.Equal(t, snapshot, got)

	requireGitlabRunnerListOk(t, runtime, name)
}

func Test_GitlabRunner_ApplySettings_WithoutRunnerEntryFails(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	runtime, name := startGitlabRunnerContainer(t, gitlabRunnerConcurrentOnly)
	provider := gitlab.New()

	dockerSettings := newPullPolicyOnlySettings()
	err := provider.ApplySettings(ctx, runtime, name, dockerSettings)
	require.ErrorIs(t, err, gitlab_runner_config.ErrRunnerEntryMissing)

	globalSettings := newGlobalOnlySettings()

	err = provider.ApplySettings(ctx, runtime, name, globalSettings)
	require.NoError(t, err)
}

func newFullRunnerSettings() domain.GitlabRunnerSettings {
	return domain.GitlabRunnerSettings{
		PullPolicy:          []string{pullAlways, pullIfNotPresent},
		AllowedPullPolicies: []string{pullAlways, pullIfNotPresent},
		CheckInterval:       7,
		LogLevel:            "debug",
		ShutdownTimeout:     30,
	}
}

func newPullPolicyOnlySettings() domain.GitlabRunnerSettings {
	return domain.GitlabRunnerSettings{
		PullPolicy: []string{pullAlways},
	}
}

func newGlobalOnlySettings() domain.GitlabRunnerSettings {
	return domain.GitlabRunnerSettings{
		CheckInterval:   7,
		LogLevel:        "debug",
		ShutdownTimeout: 30,
	}
}

func readGitlabRunnerConfig(t *testing.T, runtime container_runtime.ContainerRuntime, name string) string {
	t.Helper()

	got, err := runtime.CopyFromContainer(t.Context(), name, gitlabRunnerConfigPath)
	require.NoError(t, err)

	return string(got)
}

func requireGitlabRunnerListOk(t *testing.T, runtime container_runtime.ContainerRuntime, name string) {
	t.Helper()

	execOpts := container.ExecOptions{
		Cmd:          []string{"gitlab-runner", "list"},
		AttachStdout: true,
		AttachStderr: true,
	}

	output, exitCode, err := runtime.Exec(t.Context(), name, execOpts)
	require.NoError(t, err)
	require.Zero(t, exitCode, string(output))
	require.NotContains(t, string(output), "decoding configuration file")
	require.Contains(t, string(output), "Executor")
}
