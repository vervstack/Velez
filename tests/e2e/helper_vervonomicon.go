package e2e

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"path"
	"testing"

	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
)

const (
	vervImageDockerfile = "FROM " + HelloWorldAppImage + "\nCOPY verv/ /verv/\n"

	vervTarFileMode = 0o644
)

// buildVervImage builds a derived image, FROM HelloWorldAppImage, with files
// baked at /verv/<path> - the layout vervonomicon.ImageSource.Read expects
// (source_image.go's vervDirPath). files is keyed exactly like a Descriptor's
// Raw map: path relative to .verv/, e.g. "vervonomicon.yaml" or
// "staging/deployment.yaml".
//
// No image-build helper existed anywhere in this codebase before this one;
// authored from scratch against client.ImageBuild's in-memory tar-context
// contract (github.com/docker/docker/client). Built images are never
// explicitly removed: they only ever live in the per-test-binary DinD daemon,
// torn down wholesale at suite end (main_test.go).
func buildVervImage(t *testing.T, dockerAPI client.APIClient, tag string, files map[string]string) {
	t.Helper()

	ctx := t.Context()

	buildContext := vervBuildContextTar(t, files)

	opts := build.ImageBuildOptions{
		Tags:       []string{tag},
		Dockerfile: "Dockerfile",
		Remove:     true,
	}

	resp, err := dockerAPI.ImageBuild(ctx, buildContext, opts)
	require.NoError(t, err, "error starting image build")

	defer func() {
		closeErr := resp.Body.Close()
		require.NoError(t, closeErr)
	}()

	requireBuildSucceeded(t, resp.Body)
}

func vervBuildContextTar(t *testing.T, files map[string]string) io.Reader {
	t.Helper()

	buf := &bytes.Buffer{}
	tw := tar.NewWriter(buf)

	writeTarFile(t, tw, "Dockerfile", vervImageDockerfile)

	for relPath, content := range files {
		writeTarFile(t, tw, path.Join("verv", relPath), content)
	}

	err := tw.Close()
	require.NoError(t, err)

	return bytes.NewReader(buf.Bytes())
}

func writeTarFile(t *testing.T, tw *tar.Writer, name, content string) {
	t.Helper()

	hdr := &tar.Header{
		Name: name,
		Mode: vervTarFileMode,
		Size: int64(len(content)),
	}

	err := tw.WriteHeader(hdr)
	require.NoError(t, err)

	_, err = tw.Write([]byte(content))
	require.NoError(t, err)
}

// buildLogLine is the subset of a Docker build JSON-stream line this helper
// cares about: an "error" field means the build failed even though
// ImageBuild's own error return is nil - the daemon streams a build failure
// as a JSON message on an otherwise-200 response, not as an HTTP error.
type buildLogLine struct {
	Error string `json:"error"`
}

// createBoundResourceContainer creates a plain, never-started container named
// "<serviceName>_<resourceType>" - the naming convention
// local_storage.dockerServiceResourcesStorage.GetResources and its postgres
// equivalent both key a resources.yaml binding on (see
// internal/storage/local_storage/service_resources.go and
// internal/storage/postgres/service_resources.go). Simulates an
// already-provisioned resource without actually running one, for
// reconciliation tests.
func createBoundResourceContainer(t *testing.T, dockerAPI client.APIClient, serviceName, resourceType string) {
	t.Helper()

	ctx := t.Context()

	_, err := dockerutils.PullImage(ctx, dockerAPI, HelloWorldAppImage, false)
	require.NoError(t, err)

	containerName := serviceName + "_" + resourceType

	cfg := &container.Config{Image: HelloWorldAppImage}

	created, err := dockerAPI.ContainerCreate(ctx, cfg, nil, nil, nil, containerName)
	require.NoError(t, err)

	t.Cleanup(func() {
		removeOpts := container.RemoveOptions{Force: true}

		_ = dockerAPI.ContainerRemove(context.Background(), created.ID, removeOpts)
	})
}

func requireBuildSucceeded(t *testing.T, body io.Reader) {
	t.Helper()

	decoder := json.NewDecoder(body)

	for {
		var line buildLogLine

		err := decoder.Decode(&line)
		if errors.Is(err, io.EOF) {
			return
		}

		require.NoError(t, err, "error decoding image build log stream")

		if line.Error != "" {
			t.Fatalf("image build failed: %s", line.Error)
		}
	}
}
