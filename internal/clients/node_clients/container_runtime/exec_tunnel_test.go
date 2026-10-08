package container_runtime

import (
	"bytes"
	"io"
	"net"
	"testing"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/stretchr/testify/require"
)

func newFramedStream(t *testing.T, stdout, stderr string) *bytes.Buffer {
	t.Helper()

	stream := &bytes.Buffer{}

	_, err := stdcopy.NewStdWriter(stream, stdcopy.Stdout).Write([]byte(stdout))
	require.NoError(t, err)

	_, err = stdcopy.NewStdWriter(stream, stdcopy.Stderr).Write([]byte(stderr))
	require.NoError(t, err)

	return stream
}

func Test_DemuxConn_Read_ReturnsStdoutPayload(t *testing.T) {
	local, remote := net.Pipe()

	t.Cleanup(func() { _ = local.Close() })
	t.Cleanup(func() { _ = remote.Close() })

	stderr := &bytes.Buffer{}
	conn := newDemuxConn(local, newFramedStream(t, "HTTP/1.1 200 OK", "dial warning"), stderr)

	got, err := io.ReadAll(conn)
	require.NoError(t, err)

	require.Equal(t, "HTTP/1.1 200 OK", string(got))
	require.Equal(t, "dial warning", stderr.String())
}

func Test_DemuxConn_Write_GoesRawToConn(t *testing.T) {
	local, remote := net.Pipe()

	t.Cleanup(func() { _ = remote.Close() })

	conn := newDemuxConn(local, newFramedStream(t, "", ""), io.Discard)

	go func() {
		_, _ = conn.Write([]byte("GET /_ping"))
	}()

	buf := make([]byte, len("GET /_ping"))

	_, err := io.ReadFull(remote, buf)
	require.NoError(t, err)
	require.Equal(t, "GET /_ping", string(buf))

	require.NoError(t, conn.Close())
}
