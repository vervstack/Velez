//go:build e2e_full

package e2e

import (
	"io"
	"testing"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"
)

// ensureImagePresent pulls ref only when the daemon does not have it yet.
// A bare ImagePull always contacts the registry, which fails the test on a
// flaky Docker Hub even for images seeded into the DinD cache.
func ensureImagePresent(t *testing.T, dockerClient client.APIClient, ref string) {
	t.Helper()

	_, err := dockerClient.ImageInspect(t.Context(), ref)
	if err == nil {
		return
	}

	require.True(t, errdefs.IsNotFound(err), "error inspecting image %s: %v", ref, err)

	pullReader, err := dockerClient.ImagePull(t.Context(), ref, image.PullOptions{})
	require.NoError(t, err)

	_, err = io.Copy(io.Discard, pullReader)
	require.NoError(t, err)

	err = pullReader.Close()
	require.NoError(t, err)
}
