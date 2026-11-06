package selfswap

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"
)

const (
	testImage = "alpine:3.20"

	happyPathName    = "velez-selfswap-happy"
	rollbackPathName = "velez-selfswap-rollback"
)

// newDockerClient returns a docker client, skipping the test when no daemon
// is reachable.
func newDockerClient(t *testing.T) client.APIClient {
	t.Helper()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	_, err = cli.Ping(ctx)
	if err != nil {
		t.Skipf("docker daemon unreachable: %v", err)
	}

	t.Cleanup(func() {
		_ = cli.Close()
	})

	return cli
}

// createDummy creates a throwaway alpine sleep container and returns its id.
// The name carries the test name so parallel test binaries never collide.
func createDummy(t *testing.T, cli client.APIClient, name string) string {
	t.Helper()

	ctx := context.Background()

	_, err := cli.ImageInspect(ctx, testImage)
	if err != nil {
		pullOptions := image.PullOptions{}

		pullResp, pullErr := cli.ImagePull(ctx, testImage, pullOptions)
		require.NoError(t, pullErr)

		_, err = io.Copy(io.Discard, pullResp)
		require.NoError(t, err)
		require.NoError(t, pullResp.Close())
	}

	containerConfig := &container.Config{
		Image: testImage,
		Cmd:   []string{"sleep", "600"},
	}

	// busybox sleep as PID 1 ignores SIGTERM and every stop would burn the
	// full stop timeout; tini forwards the signal so stops are instant.
	runInit := true
	hostConfig := &container.HostConfig{
		Init: &runInit,
	}

	created, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, name)
	require.NoError(t, err)

	err = cli.ContainerStart(ctx, created.ID, container.StartOptions{})
	require.NoError(t, err)

	t.Cleanup(func() {
		removeOptions := container.RemoveOptions{Force: true}

		_ = cli.ContainerRemove(context.Background(), name, removeOptions)
		_ = cli.ContainerRemove(context.Background(), RollbackContainerName, removeOptions)
	})

	return created.ID
}

func newTestSwapper(cli client.APIClient, targetId string, healthErr error) *swapper {
	return &swapper{
		cli: cli,
		log: discardLogger(),
		settings: settings{
			targetContainerId: targetId,
			healthTimeout:     time.Second,
			rollbackWindow:    time.Millisecond * 200,
		},
		checkHealth: func(_ context.Context, _ string, _ time.Duration) error {
			return healthErr
		},
	}
}

// Test_run_happyPath swaps a throwaway container for a clone of itself: the
// original name ends up on a running container, the rollback container is
// dropped after its window.
func Test_run_happyPath(t *testing.T) {
	cli := newDockerClient(t)
	ctx := context.Background()

	targetId := createDummy(t, cli, happyPathName)

	sw := newTestSwapper(cli, targetId, nil)

	err := sw.run(ctx)
	require.NoError(t, err)

	newInspect, err := cli.ContainerInspect(ctx, happyPathName)
	require.NoError(t, err)
	require.True(t, newInspect.State.Running, "replacement must be running")
	require.NotEqual(t, targetId, newInspect.ID, "replacement must be a new container")
	require.Equal(t, testImage, newInspect.Config.Image)

	_, err = cli.ContainerInspect(ctx, RollbackContainerName)
	require.Error(t, err, "rollback container must be dropped after the window")
}

// Test_run_rollbackPath fails the health check and verifies the old
// container is restored under its original name and running.
func Test_run_rollbackPath(t *testing.T) {
	cli := newDockerClient(t)
	ctx := context.Background()

	targetId := createDummy(t, cli, rollbackPathName)

	healthErr := rerrors.New("simulated health check failure")
	sw := newTestSwapper(cli, targetId, healthErr)

	err := sw.run(ctx)
	require.ErrorIs(t, err, healthErr)

	restored, err := cli.ContainerInspect(ctx, rollbackPathName)
	require.NoError(t, err)
	require.Equal(t, targetId, restored.ID, "original container must be restored")
	require.True(t, restored.State.Running, "restored container must be running")

	_, err = cli.ContainerInspect(ctx, RollbackContainerName)
	require.Error(t, err, "rollback name must be released after rename back")
}
