package container_runtime

import (
	"context"
	"io"
	"net"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"go.redsock.ru/rerrors"
)

// ExecTunnelDialContext opens a connection to the Docker daemon running inside
// a container (a DinD).
type ExecTunnelDialContext func(ctx context.Context, network, addr string) (net.Conn, error)

// NewExecTunnelDialer tunnels connections to the daemon inside containerId
// through `docker system dial-stdio` exec'd there over cli, the outer daemon:
// no network join and no published port, so it works wherever cli itself is
// reachable.
func NewExecTunnelDialer(cli client.APIClient, containerId string) ExecTunnelDialContext {
	return func(ctx context.Context, _, _ string) (net.Conn, error) {
		execOptions := container.ExecOptions{
			Cmd:          []string{"docker", "system", "dial-stdio"},
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
		}

		created, err := cli.ContainerExecCreate(ctx, containerId, execOptions)
		if err != nil {
			return nil, rerrors.Wrap(err, "error creating dial-stdio exec")
		}

		hijacked, err := cli.ContainerExecAttach(ctx, created.ID, container.ExecAttachOptions{})
		if err != nil {
			return nil, rerrors.Wrap(err, "error attaching to dial-stdio exec")
		}

		return newDemuxConn(hijacked.Conn, hijacked.Reader, io.Discard), nil
	}
}

// demuxConn is a net.Conn over an attached exec without a TTY: writes go raw to
// the exec's stdin, reads come out of the stdcopy-multiplexed output stream.
type demuxConn struct {
	conn   net.Conn
	reader *io.PipeReader

	closeOnce sync.Once
}

// newDemuxConn splits the multiplexed output in source into the payload that
// Read returns; stderr frames go to stderr.
func newDemuxConn(conn net.Conn, source io.Reader, stderr io.Writer) *demuxConn {
	reader, writer := io.Pipe()

	go func() {
		_, err := stdcopy.StdCopy(writer, stderr, source)

		_ = writer.CloseWithError(err)
	}()

	return &demuxConn{conn: conn, reader: reader}
}

// The net.Conn methods return their errors as they are: net/http compares
// io.EOF and net errors directly, and wrapping would break that.

func (d *demuxConn) Read(p []byte) (int, error) {
	return d.reader.Read(p) //nolint:wrapcheck // net.Conn contract, see above
}

func (d *demuxConn) Write(p []byte) (int, error) {
	return d.conn.Write(p) //nolint:wrapcheck // net.Conn contract, see above
}

func (d *demuxConn) Close() error {
	var err error

	d.closeOnce.Do(func() {
		_ = d.reader.Close()
		err = d.conn.Close()
	})

	return err //nolint:wrapcheck // net.Conn contract, see above
}

func (d *demuxConn) LocalAddr() net.Addr {
	return d.conn.LocalAddr()
}

func (d *demuxConn) RemoteAddr() net.Addr {
	return d.conn.RemoteAddr()
}

func (d *demuxConn) SetDeadline(t time.Time) error {
	return d.conn.SetDeadline(t) //nolint:wrapcheck // net.Conn contract, see above
}

func (d *demuxConn) SetReadDeadline(t time.Time) error {
	return d.conn.SetReadDeadline(t) //nolint:wrapcheck // net.Conn contract, see above
}

func (d *demuxConn) SetWriteDeadline(t time.Time) error {
	return d.conn.SetWriteDeadline(t) //nolint:wrapcheck // net.Conn contract, see above
}
