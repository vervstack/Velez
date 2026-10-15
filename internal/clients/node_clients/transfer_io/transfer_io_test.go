package transfer_io_test

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/clients/node_clients/transfer_io"
)

type testDocker struct {
	cli client.APIClient
}

func (d testDocker) Client() client.APIClient {
	return d.cli
}

func newTransferIO(t *testing.T) (transfer_io.TransferIO, client.APIClient) {
	t.Helper()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = cli.Close()
	})

	_, err = cli.Ping(t.Context())
	if err != nil {
		t.Skip("docker daemon is unreachable")
	}

	return transfer_io.New(testDocker{cli: cli}), cli
}

func newVolumeName(t *testing.T, cli client.APIClient) string {
	t.Helper()

	name := "velez_transfer_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")

	t.Cleanup(func() {
		_ = cli.VolumeRemove(context.WithoutCancel(t.Context()), name, true)
	})

	return name
}

func newVolumeTar(t *testing.T, files map[string]string) io.Reader {
	t.Helper()

	var buf bytes.Buffer

	writer := tar.NewWriter(&buf)

	for name, content := range files {
		header := tar.Header{
			Name:     name,
			Mode:     0o644,
			Size:     int64(len(content)),
			Typeflag: tar.TypeReg,
		}

		err := writer.WriteHeader(&header)
		require.NoError(t, err)

		_, err = writer.Write([]byte(content))
		require.NoError(t, err)
	}

	err := writer.Close()
	require.NoError(t, err)

	return &buf
}

func readVolumeTar(t *testing.T, stream io.Reader) map[string]string {
	t.Helper()

	files := make(map[string]string)
	reader := tar.NewReader(stream)

	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			return files
		}

		require.NoError(t, err)

		if header.Typeflag != tar.TypeReg {
			continue
		}

		content, err := io.ReadAll(reader)
		require.NoError(t, err)

		files[strings.TrimPrefix(header.Name, "./")] = string(content)
	}
}

func Test_VolumeWriteRead_RoundTrip(t *testing.T) {
	t.Parallel()

	transfer, cli := newTransferIO(t)
	volumeName := newVolumeName(t, cli)

	want := map[string]string{
		"a.txt":         "hello transfer",
		"sub/b/deep.md": "nested content",
	}

	err := transfer.VolumeWrite(t.Context(), volumeName, newVolumeTar(t, want))
	require.NoError(t, err)

	stream, err := transfer.VolumeRead(t.Context(), volumeName)
	require.NoError(t, err)

	got := readVolumeTar(t, stream)

	err = stream.Close()
	require.NoError(t, err)

	require.Equal(t, want, got)

	size, err := transfer.VolumeSize(t.Context(), volumeName)
	require.NoError(t, err)
	require.Positive(t, size)
}

func Test_VolumeWrite_ExistingVolume(t *testing.T) {
	t.Parallel()

	transfer, cli := newTransferIO(t)
	volumeName := newVolumeName(t, cli)

	createOpts := volume.CreateOptions{Name: volumeName}

	_, err := cli.VolumeCreate(t.Context(), createOpts)
	require.NoError(t, err)

	err = transfer.VolumeWrite(t.Context(), volumeName, newVolumeTar(t, map[string]string{"a.txt": "x"}))
	require.Error(t, err)
	require.Contains(t, err.Error(), "volume already exists")
}

func Test_ImageSaveLoad_RoundTrip(t *testing.T) {
	t.Parallel()

	transfer, cli := newTransferIO(t)

	pull, err := cli.ImagePull(t.Context(), "alpine", image.PullOptions{})
	require.NoError(t, err)

	_, err = io.Copy(io.Discard, pull)
	require.NoError(t, err)

	err = pull.Close()
	require.NoError(t, err)

	saved, err := transfer.ImageSave(t.Context(), "alpine")
	require.NoError(t, err)

	err = transfer.ImageLoad(t.Context(), saved)
	require.NoError(t, err)

	err = saved.Close()
	require.NoError(t, err)
}
