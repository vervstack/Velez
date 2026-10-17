//go:build e2e_full

package e2e

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/clients/node_clients/local_state"
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
	env.Cfg = prev.Cfg
	env.velezKey = prev.velezKey
	env.velez = prev.velez

	connectVelezGrpc(t, &env)

	return &env
}

// restartWithState stops the Velez container, rewrites its state file through
// mutate and starts it again: the settings Velez reads at startup change, its
// managed containers and the gRPC connection stay.
func (e *TestEnvironment) restartWithState(t *testing.T, mutate func(state *local_state.State)) {
	t.Helper()

	var state local_state.State

	err := json.Unmarshal(e.containerStateFile(t), &state)
	require.NoError(t, err)

	mutate(&state)

	seeded, err := json.Marshal(state)
	require.NoError(t, err)

	grace := velezRestartGrace
	stopOpts := container.StopOptions{Timeout: &grace}

	err = e.DockerClient().ContainerStop(t.Context(), e.velez.id, stopOpts)
	require.NoError(t, err)

	var buf bytes.Buffer

	writer := tar.NewWriter(&buf)

	addTarFile(t, writer, strings.TrimPrefix(velezStateDir, "/")+"/"+velezStateFile, seeded)

	err = writer.Close()
	require.NoError(t, err)

	copyOpts := container.CopyToContainerOptions{}

	err = e.DockerClient().CopyToContainer(t.Context(), e.velez.id, "/", &buf, copyOpts)
	require.NoError(t, err)

	err = e.DockerClient().ContainerStart(t.Context(), e.velez.id, container.StartOptions{})
	require.NoError(t, err)

	awaitVelezReady(t, e)
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
