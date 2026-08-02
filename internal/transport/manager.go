package transport

import (
	"context"
	"net"
	"net/http"
	"sync/atomic"

	"github.com/rs/zerolog/log"
	"github.com/soheilhy/cmux"
	"go.redsock.ru/rerrors"
	"golang.org/x/sync/errgroup"
)

// noCloseListener wraps a net.Listener and turns Close into a no-op.
//
// grpc.Server.Serve and net/http's Server.Serve both close the listener they
// were handed once Serve returns - documented behavior in both libraries, not
// specific to a graceful-vs-hard stop. cmux's matched listeners (from
// CMux.Match) embed the shared root net.Listener directly without overriding
// Close, so letting either server's Serve call through would tear down the
// root listener - and with it every other protocol sharing the mux - out from
// under ServersManager. Root listener lifecycle belongs solely to
// ServersManager.Stop.
type noCloseListener struct {
	net.Listener
}

func (noCloseListener) Close() error { return nil }

type ServersManager struct {
	grpcServer
	httpServer

	mux          cmux.CMux
	rootListener net.Listener

	stopping atomic.Bool
}

func NewServerManager(ctx context.Context, listener net.Listener) (*ServersManager, error) {
	mainMux := cmux.New(listener)
	httpMux := http.NewServeMux()

	s := &ServersManager{
		mux:          mainMux,
		rootListener: listener,

		grpcServer: newGrpcServer(noCloseListener{mainMux.Match(cmux.HTTP2())}, httpMux),
		httpServer: newHttpServer(noCloseListener{mainMux.Match(cmux.Any())}, httpMux),
	}

	return s, nil
}

func (m *ServersManager) Start() error {
	log.Info().Msg("Starting server at http://0.0.0.0" + m.grpcServer.listener.Addr().String()[4:])

	errGroup, _ := errgroup.WithContext(context.Background())

	errGroup.Go(func() error {
		err := m.mux.Serve()
		if err != nil && !m.stopping.Load() {
			return rerrors.Wrap(err, "mux failed to serve")
		}

		return nil
	})
	errGroup.Go(m.grpcServer.start)
	errGroup.Go(m.httpServer.start)

	err := errGroup.Wait()
	if err != nil {
		return rerrors.Wrap(err, "error returned from errgroup when starting servers ")
	}

	return nil
}

func (m *ServersManager) Stop() error {
	m.stopping.Store(true)

	errGroup, _ := errgroup.WithContext(context.Background())

	errGroup.Go(m.grpcServer.stop)
	errGroup.Go(m.httpServer.stop)
	errGroup.Go(func() error {
		// Close signals the matched listeners first (cmux's own
		// ErrServerClosed/ErrListenerClosed, so grpcServer.start/httpServer.start
		// return cleanly), then the root listener unblocks CMux.Serve's own
		// Accept loop above.
		m.mux.Close()

		return m.rootListener.Close()
	})

	err := errGroup.Wait()
	if err != nil {
		return rerrors.Wrap(err, "error stopping server")
	}

	return nil
}
