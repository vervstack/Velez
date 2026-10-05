//go:build e2e_full

package e2e

import (
	"io"
	"net"
	"strconv"
	"sync"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
)

const loopbackHost = "127.0.0.1"

var dindLoopbackOnce sync.Once //nolint:gochecknoglobals // one bridge per test binary

// requireDindLoopbackBridge makes "localhost:<port>" on this host reach the
// DinD-side port <port> for the whole allocator band.
//
// A Velez binary that talks to a container it published a port for dials
// localhost:<published port>, which is right when the Docker daemon shares the
// host's network namespace. Under the DinD harness the same port is only
// reachable through an ephemeral bootstrap-host port (sharedDind.Addr), so the
// product's own address would otherwise miss. A port already held by another
// process is skipped; a test that is handed such a port fails on its own.
func requireDindLoopbackBridge(t *testing.T) {
	t.Helper()

	require.NotNil(t, sharedDind, "dind harness is not up")

	dindLoopbackOnce.Do(func() {
		for port := dindPortBandStart; port <= dindAllocatorEnd; port++ {
			listener, err := net.Listen("tcp", net.JoinHostPort(loopbackHost, strconv.Itoa(port)))
			if err != nil {
				log.Warn().Err(err).Int("port", port).Msg("dind loopback bridge: port is busy, skipped")

				continue
			}

			go serveLoopbackBridge(listener, port)
		}
	})
}

func serveLoopbackBridge(listener net.Listener, dindPort int) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}

		go forwardToDind(conn, dindPort)
	}
}

func forwardToDind(client net.Conn, dindPort int) {
	defer func() { _ = client.Close() }()

	target, ok := sharedDind.Addr(dindPort)
	if !ok {
		return
	}

	upstream, err := net.Dial("tcp", target)
	if err != nil {
		return
	}

	defer func() { _ = upstream.Close() }()

	done := make(chan struct{}, 2)

	go func() {
		_, _ = io.Copy(upstream, client)
		done <- struct{}{}
	}()

	go func() {
		_, _ = io.Copy(client, upstream)
		done <- struct{}{}
	}()

	<-done
}
