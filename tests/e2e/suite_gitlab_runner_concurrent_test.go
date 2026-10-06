//go:build e2e_full

package e2e

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker"
	"go.vervstack.ru/Velez/internal/service/service_manager/runneraas/providers/gitlab"
)

const (
	gitlabRunnerImage      = "gitlab/gitlab-runner:latest"
	gitlabRunnerConfigPath = "/etc/gitlab-runner/config.toml"
	gitlabRunnerSeedConfig = "concurrent = 1\n[[runners]]\n  name = \"x\"\n"
)

func Test_GitlabRunner_ApplyConcurrent_EditsConfigToml(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	runtime, name := startGitlabRunnerContainer(t, gitlabRunnerSeedConfig)

	err := gitlab.New().ApplyConcurrent(ctx, runtime, name, 4)
	require.NoError(t, err)

	got, err := runtime.CopyFromContainer(ctx, name, gitlabRunnerConfigPath)
	require.NoError(t, err)

	require.Equal(t, "concurrent = 4\n[[runners]]\n  name = \"x\"\n", string(got))
}

func startGitlabRunnerContainer(t *testing.T, seed string) (container_runtime.ContainerRuntime, string) {
	t.Helper()

	ctx := t.Context()
	name := GetServiceName(t)

	dockerClient, err := docker.NewClient(nil)
	require.NoError(t, err)

	api := dockerClient.Client()

	runtimes := container_runtime.NewResolver(api, dockerClient.Host(), nil, nil)

	runtime, err := runtimes.Runtime(ctx, "")
	require.NoError(t, err)

	_, err = runtime.PullImage(ctx, gitlabRunnerImage)
	require.NoError(t, err)

	cfg := &container.Config{
		Image:      gitlabRunnerImage,
		Entrypoint: []string{"sleep", "infinity"},
	}

	created, err := api.ContainerCreate(ctx, cfg, nil, nil, nil, name)
	require.NoError(t, err)

	t.Cleanup(func() {
		removeOpts := container.RemoveOptions{Force: true}

		_ = api.ContainerRemove(context.Background(), created.ID, removeOpts)
	})

	err = api.ContainerStart(ctx, created.ID, container.StartOptions{})
	require.NoError(t, err)

	err = runtime.CopyToContainer(ctx, name, gitlabRunnerConfigPath, []byte(seed), 0o600)
	require.NoError(t, err)

	return runtime, name
}
