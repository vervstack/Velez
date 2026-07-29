package transport

import (
	"context"
	"net"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/soheilhy/cmux"
	"go.redsock.ru/rerrors"
	"golang.org/x/sync/errgroup"
)

type ServersManager struct {
	grpcServer
	httpServer

	mux cmux.CMux
}

func NewServerManager(ctx context.Context, listener net.Listener) (*ServersManager, error) {
	mainMux := cmux.New(listener)
	httpMux := http.NewServeMux()

	s := &ServersManager{
		mux: mainMux,

		grpcServer: newGrpcServer(ctx, mainMux.Match(cmux.HTTP2()), httpMux),
		httpServer: newHTTPServer(mainMux.Match(cmux.Any()), httpMux),
	}

	return s, nil
}

func (m *ServersManager) Start() error {
	log.Info().Msg("Starting server at http://0.0.0.0" + m.grpcServer.listener.Addr().String()[4:])

	errGroup, ctx := errgroup.WithContext(context.Background())

	errGroup.Go(func() error {
		err := m.mux.Serve()
		if err != nil {
			return rerrors.Wrap(err, "mux faild to serve")
		}

		return nil
	})
	errGroup.Go(func() error {
		err := m.grpcServer.start()
		if err != nil {
			return rerrors.Wrap(err, "failed to start grpc server")
		}

		return nil
	})
	errGroup.Go(func() error {
		err := m.httpServer.start()
		if err != nil {
			return rerrors.Wrap(err, "failed to start http server")
		}

		return nil
	})

	errC := make(chan error, 1)

	select {
	case <-ctx.Done():
		return nil
	case errC <- errGroup.Wait():
		err := <-errC

		return rerrors.Wrap(err, "received error via channel in server manager Start func")
	}
}

func (m *ServersManager) Stop() error {
	eg, _ := errgroup.WithContext(context.Background())

	eg.Go(m.grpcServer.stop)
	eg.Go(m.httpServer.stop)
	eg.Go(func() error {
		m.mux.Close()

		return nil
	})

	err := eg.Wait()
	if err != nil {
		return rerrors.Wrap(err, "error stopping server")
	}

	return nil
}
