package dind

import (
	"archive/tar"
	"bytes"
	"context"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"go.redsock.ru/rerrors"
)

const (
	// EnvSocksBin is the path of a linux socksproxy binary built for the
	// bootstrap daemon's architecture. Set, Setup runs it inside the DinD and
	// publishes its port (see Env.SocksAddr); unset, the DinD has no proxy.
	EnvSocksBin = "VELEZ_E2E_SOCKS_BIN"

	socksPort         = 1080
	socksBinInDind    = "/socksproxy"
	socksBinMode      = 0o755
	socksProbeTimeout = time.Second
	socksReadyWait    = 10 * time.Second
	socksReadyPoll    = 200 * time.Millisecond
	socksReplyLen     = 2
)

//nolint:forbidigo // package-private test-infra sentinel, not shared/user-facing
var errSocksNotReady = rerrors.New("socks proxy did not become ready")

// SocksAddr returns the bootstrap-host address of the SOCKS5 proxy running in
// the DinD's network namespace, and whether the DinD has one.
func (e *Env) SocksAddr() (string, bool) {
	return e.Addr(socksPort)
}

func socksBinPath() string {
	return os.Getenv(EnvSocksBin)
}

func withSocksPort(publish []int) []int {
	if socksBinPath() == "" {
		return publish
	}

	withProxy := make([]int, 0, len(publish)+1)

	withProxy = append(withProxy, publish...)

	return append(withProxy, socksPort)
}

// startSocksProxy copies the proxy binary into the DinD and starts it, unless
// one already answers on the published port.
func (e *Env) startSocksProxy(ctx context.Context) error {
	binPath := socksBinPath()
	if binPath == "" {
		return nil
	}

	addr, ok := e.SocksAddr()
	if !ok {
		return rerrors.Wrap(errSocksNotReady, "socks port was not published")
	}

	if isSocksUp(addr) {
		return nil
	}

	binary, err := os.ReadFile(binPath)
	if err != nil {
		return rerrors.Wrap(err, "error reading socks proxy binary")
	}

	archive, err := singleFileTar(socksBinInDind[1:], binary)
	if err != nil {
		return rerrors.Wrap(err, "error packing socks proxy binary")
	}

	copyOpts := container.CopyToContainerOptions{}

	err = e.bootstrap.CopyToContainer(ctx, e.containerId, "/", archive, copyOpts)
	if err != nil {
		return rerrors.Wrap(err, "error copying socks proxy into dind")
	}

	err = e.execDetached(ctx, socksBinInDind, "-listen", ":"+strconv.Itoa(socksPort))
	if err != nil {
		return rerrors.Wrap(err, "error starting socks proxy")
	}

	return waitForSocks(ctx, addr)
}

func (e *Env) execDetached(ctx context.Context, cmd ...string) error {
	return execDetached(ctx, e.bootstrap, e.containerId, cmd...)
}

func execDetached(ctx context.Context, cli *client.Client, containerId string, cmd ...string) error {
	execOpts := container.ExecOptions{Cmd: cmd}

	created, err := cli.ContainerExecCreate(ctx, containerId, execOpts)
	if err != nil {
		return rerrors.Wrap(err, "error creating exec")
	}

	startOpts := container.ExecStartOptions{Detach: true}

	err = cli.ContainerExecStart(ctx, created.ID, startOpts)
	if err != nil {
		return rerrors.Wrap(err, "error starting exec")
	}

	return nil
}

func singleFileTar(name string, content []byte) (*bytes.Reader, error) {
	var buf bytes.Buffer

	writer := tar.NewWriter(&buf)

	header := &tar.Header{
		Name: name,
		Mode: socksBinMode,
		Size: int64(len(content)),
	}

	err := writer.WriteHeader(header)
	if err != nil {
		return nil, rerrors.Wrap(err, "error writing tar header")
	}

	_, err = writer.Write(content)
	if err != nil {
		return nil, rerrors.Wrap(err, "error writing tar body")
	}

	err = writer.Close()
	if err != nil {
		return nil, rerrors.Wrap(err, "error closing tar")
	}

	return bytes.NewReader(buf.Bytes()), nil
}

// isSocksUp performs a SOCKS5 greeting; a bare TCP accept is not enough, the
// docker port forwarder accepts before anything listens behind it.
func isSocksUp(addr string) bool {
	dialer := &net.Dialer{Timeout: socksProbeTimeout}

	conn, err := dialer.DialContext(context.Background(), "tcp", addr)
	if err != nil {
		return false
	}

	defer func() { _ = conn.Close() }()

	_ = conn.SetDeadline(time.Now().Add(socksProbeTimeout))

	_, err = conn.Write([]byte{5, 1, 0})
	if err != nil {
		return false
	}

	reply := make([]byte, socksReplyLen)

	_, err = conn.Read(reply)

	return err == nil && reply[0] == 5 && reply[1] == 0
}

func waitForSocks(ctx context.Context, addr string) error {
	deadline := time.Now().Add(socksReadyWait)

	for time.Now().Before(deadline) {
		if isSocksUp(addr) {
			return nil
		}

		select {
		case <-ctx.Done():
			return rerrors.Wrap(ctx.Err(), "error waiting for socks proxy")
		case <-time.After(socksReadyPoll):
		}
	}

	return rerrors.Wrap(errSocksNotReady)
}
