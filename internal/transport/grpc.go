package transport

import (
	"context"
	"net"
	"net/http"

	"github.com/soheilhy/cmux"
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcImpl interface {
	Register(srv grpc.ServiceRegistrar)
}

type GrpcWithGateway interface {
	Gateway(ctx context.Context, endpoint string, opts ...grpc.DialOption) (rootRoute string, handler http.Handler)
}

type grpcServer struct {
	listener net.Listener

	gatewayMux *http.ServeMux

	opts            []grpc.ServerOption
	implementations []GrpcImpl

	// AvailableAfter start is called
	stopCall func()
}

func newGrpcServer(
	listener net.Listener,
	gatewayMux *http.ServeMux) grpcServer {
	return grpcServer{
		listener:   listener,
		stopCall:   func() {},
		gatewayMux: gatewayMux,
	}
}

func (s *grpcServer) AddImplementation(ctx context.Context, grpcImpls ...GrpcImpl) {
	for _, grpcImpl := range grpcImpls {
		s.implementations = append(s.implementations, grpcImpl)

		grpcWithGateway, ok := grpcImpl.(GrpcWithGateway)
		if ok {
			s.gatewayMux.Handle(grpcWithGateway.Gateway(ctx,
				s.listener.Addr().String(),
				grpc.WithTransportCredentials(insecure.NewCredentials())))
		}
	}
}

func (s *grpcServer) AddServerOption(opts ...grpc.ServerOption) {
	s.opts = append(s.opts, opts...)
}

func (s *grpcServer) start() error {
	server := grpc.NewServer(s.opts...)

	for _, impl := range s.implementations {
		impl.Register(server)
	}

	// stopCall is intentionally NOT set to server.GracefulStop: cmux's matched
	// listener embeds the shared root net.Listener without overriding Close(),
	// so calling GracefulStop/Stop here would close the root listener out from
	// under the whole mux (grpc and http share one socket). Shutdown goes
	// through ServersManager.Stop's m.mux.Close() instead, which unblocks
	// Serve() via cmux's own ErrServerClosed/ErrListenerClosed signal without
	// touching the socket.
	err := server.Serve(s.listener)
	if err != nil {
		if !rerrors.Is(err, cmux.ErrServerClosed) && !rerrors.Is(err, cmux.ErrListenerClosed) {
			return rerrors.Wrap(err, "error serving grpc server")
		}
	}

	return nil
}

func (s *grpcServer) stop() error {
	s.stopCall()

	return nil
}
