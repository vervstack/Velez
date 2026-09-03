package dockerutils_test

import (
	"archive/tar"
	"bytes"
	"testing"
	"time"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
	"go.vervstack.ru/Velez/tests/test_helper"
)

// tarEntry describes one entry to bake into a test tar archive via
// buildTar - a regular file, a directory, or a symlink, depending on
// typeflag.
type tarEntry struct {
	name     string
	typeflag byte
	content  string
	linkname string
}

// buildTar packs entries into an in-memory tar archive suitable for
// client.APIClient.CopyToContainer, so a test can seed a container's
// filesystem without shelling out to exec.
func buildTar(t *testing.T, entries []tarEntry) *bytes.Buffer {
	t.Helper()

	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	for _, entry := range entries {
		hdr := &tar.Header{
			Name:     entry.name,
			Typeflag: entry.typeflag,
			Mode:     0o644,
			Size:     int64(len(entry.content)),
			Linkname: entry.linkname,
			ModTime:  time.Now(),
		}

		if entry.typeflag == tar.TypeDir {
			hdr.Mode = 0o755
			hdr.Size = 0
		}

		err := tw.WriteHeader(hdr)
		require.NoError(t, err)

		_, err = tw.Write([]byte(entry.content))
		require.NoError(t, err)
	}

	err := tw.Close()
	require.NoError(t, err)

	return buf
}

// newTestContainer creates (but does not start) a throwaway container from
// the shared hello_world fixture image, registering its removal on cleanup.
// docker cp works against a created-but-not-started container, mirroring
// internal/jobs's scratch-container pattern, so no start/wait is needed.
func newTestContainer(t *testing.T, cli client.APIClient) string {
	t.Helper()

	ctx := t.Context()

	test_helper.EnsurePulled(t, cli, test_helper.HelloWorldAppImage)

	name := test_helper.UniqueName(t, "dockerutils_readdir")

	cfg := &container.Config{
		Image: test_helper.HelloWorldAppImage,
	}
	hostCfg := &container.HostConfig{}

	created, err := cli.ContainerCreate(ctx, cfg, hostCfg, nil, nil, name)
	require.NoError(t, err)

	t.Cleanup(func() {
		test_helper.RemoveContainer(t, cli, created.ID)
	})

	return created.ID
}

func Test_ReadDirFromContainer_ReturnsRegularFilesKeyedByRelativePathSkippingDirsAndSymlinks(t *testing.T) {
	t.Parallel()

	cli := test_helper.NewRealDockerAPI(t)
	contId := newTestContainer(t, cli)

	entries := []tarEntry{
		{name: "verv/", typeflag: tar.TypeDir},
		{name: "verv/vervonomicon.yaml", typeflag: tar.TypeReg, content: "root: true\n"},
		{name: "verv/prod/", typeflag: tar.TypeDir},
		{name: "verv/prod/ingress.conf", typeflag: tar.TypeReg, content: "listen 80;\n"},
		{name: "verv/link", typeflag: tar.TypeSymlink, linkname: "vervonomicon.yaml"},
	}

	tarBuf := buildTar(t, entries)

	ctx := t.Context()

	copyOpts := container.CopyToContainerOptions{}

	err := cli.CopyToContainer(ctx, contId, "/", tarBuf, copyOpts)
	require.NoError(t, err)

	got, err := dockerutils.ReadDirFromContainer(ctx, cli, contId, "/verv")
	require.NoError(t, err)

	want := map[string][]byte{
		"vervonomicon.yaml": []byte("root: true\n"),
		"prod/ingress.conf": []byte("listen 80;\n"),
	}
	require.Equal(t, want, got)
}

func Test_ReadDirFromContainer_MissingPathIsClassifiedAsNotFound(t *testing.T) {
	t.Parallel()

	cli := test_helper.NewRealDockerAPI(t)
	contId := newTestContainer(t, cli)

	ctx := t.Context()

	_, err := dockerutils.ReadDirFromContainer(ctx, cli, contId, "/does/not/exist")
	require.Error(t, err)
	require.True(t, errdefs.IsNotFound(err))
}
