//go:build e2e_full

package e2e

import (
	"archive/tar"
	"io"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/app"
)

// restartedContainerEnvironment restarts the Velez container in place: same
// state file, same managed containers, a fresh process.
func restartedContainerEnvironment(t *testing.T, prev *TestEnvironment) *TestEnvironment {
	t.Helper()

	grace := velezRestartGrace
	stopOpts := container.StopOptions{Timeout: &grace}

	err := prev.DockerClient().ContainerRestart(t.Context(), prev.velez.id, stopOpts)
	require.NoError(t, err)

	var env TestEnvironment

	env.t = t
	env.App = app.App{Ctx: t.Context(), Cfg: prev.Cfg}
	env.velezKey = prev.velezKey
	env.velez = prev.velez
	env.isRestart = true

	connectVelezGrpc(t, &env)

	return &env
}

// containerFile reads one file out of the Velez container.
func (e *TestEnvironment) containerFile(t *testing.T, path string) []byte {
	t.Helper()

	archive, _, err := e.DockerClient().CopyFromContainer(t.Context(), e.velez.id, path)
	require.NoError(t, err)

	defer func() { _ = archive.Close() }()

	reader := tar.NewReader(archive)

	_, err = reader.Next()
	require.NoError(t, err)

	content, err := io.ReadAll(reader)
	require.NoError(t, err)

	return content
}

func (e *TestEnvironment) containerStateFile(t *testing.T) []byte {
	t.Helper()

	return e.containerFile(t, velezStateDir+"/"+velezStateFile)
}

// dindReservedPorts is the reserved block tests bind themselves; the container
// backend draws the shared test pool from it because the allocator range is
// split into per-Velez slots.
func dindReservedPorts() []int {
	ports := make([]int, 0, dindReservedPortCount)

	for p := dindAllocatorEnd + 1; p <= dindPortBandEnd; p++ {
		ports = append(ports, p)
	}

	return ports
}
