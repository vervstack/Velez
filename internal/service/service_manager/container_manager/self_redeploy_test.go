package container_manager

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/selfswap"
)

func newDockerClientOrSkip(t *testing.T) client.APIClient {
	t.Helper()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	require.NoError(t, err)

	t.Cleanup(func() { _ = cli.Close() })

	_, err = cli.Ping(t.Context())
	if err != nil {
		t.Skip("docker daemon is unreachable")
	}

	return cli
}

// The swapper name is a fixed node-wide constant, so this test must not run
// in parallel with another swapper-creating test.
func Test_SelfRedeploy_CreatesSwapperContainer(t *testing.T) {
	cli := newDockerClientOrSkip(t)

	svc := &ContainerServiceImpl{dockerAPI: cli}

	removeOptions := container.RemoveOptions{Force: true}

	t.Cleanup(func() {
		_ = cli.ContainerRemove(context.WithoutCancel(t.Context()), selfswap.SwapperContainerName, removeOptions)
	})

	const testImage = "alpine:3.20"

	const fakeTargetId = "fake-target-id"

	err := svc.SelfRedeploy(t.Context(), testImage, fakeTargetId)
	require.NoError(t, err)

	inspect, err := cli.ContainerInspect(t.Context(), selfswap.SwapperContainerName)
	require.NoError(t, err)

	require.Equal(t, testImage, inspect.Config.Image)
	require.Equal(t, map[string]string{labels.SwapperLabel: fakeTargetId}, inspect.Config.Labels)
	require.Contains(t, inspect.Config.Env, selfswap.EnvSelfSwap+"=1")
	require.Contains(t, inspect.Config.Env, selfswap.EnvTargetContainerId+"="+fakeTargetId)
	require.Contains(t, inspect.Config.Env, selfswap.EnvImage+"="+testImage)
	require.Contains(t, inspect.Config.Env, selfswap.EnvHealthTimeout+"=60")
	require.Contains(t, inspect.Config.Env, selfswap.EnvRollbackWindow+"=600")
	require.Contains(t, inspect.HostConfig.Binds, "/var/run/docker.sock:/var/run/docker.sock")
	require.Equal(t, container.RestartPolicyDisabled, inspect.HostConfig.RestartPolicy.Name)

	// A second run must replace a stale swapper instead of name-conflicting.
	err = svc.SelfRedeploy(t.Context(), testImage, fakeTargetId)
	require.NoError(t, err)
}
