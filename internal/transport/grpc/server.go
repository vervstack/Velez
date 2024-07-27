// Code generated by RedSock CLI

package grpc

import (
	"context"
	"net"
	"net/http"
	"strings"
	"sync"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/godverv/matreshka/servers"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sirupsen/logrus"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/godverv/Velez/internal/clients/security"
	"github.com/godverv/Velez/internal/config"
	"github.com/godverv/Velez/internal/service"
	"github.com/godverv/Velez/pkg/velez_api"
)

type Server struct {
	*Api

	serverMux cmux.CMux

	grpcServer *grpc.Server

	serverAddress string
	m             sync.Mutex
}

func NewServer(
	cfg config.Config,
	grpcServer *servers.GRPC,

	srv service.Services,
	secManager security.Manager,
) (*Server, error) {

	var opts []grpc.ServerOption

	if !cfg.GetEnvironment().DisableAPISecurity {
		opts = append(opts, security.GrpcInterceptor(secManager))
	}

	server := &Server{
		Api:           NewApi(cfg, srv),
		grpcServer:    grpc.NewServer(opts...),
		serverAddress: "0.0.0.0:" + grpcServer.GetPortStr(),
	}

	velez_api.RegisterVelezAPIServer(server.grpcServer, server)

	return server, nil
}

func (s *Server) Start(_ context.Context) error {
	s.m.Lock()
	defer s.m.Unlock()

	listener, err := net.Listen("tcp", s.serverAddress)
	if err != nil {
		return errors.Wrap(err, "error opening listener")
	}
	s.serverMux = cmux.New(listener)

	go s.startGRPC()

	go s.startGateway()

	go func() {
		serveErr := s.serverMux.Serve()
		if serveErr != nil {
			if !strings.Contains(serveErr.Error(), "closed network connection") {
				logrus.Errorf("error service server %s", serveErr)
			}
		}
	}()

	return nil
}

func (s *Server) Stop(_ context.Context) error {
	s.grpcServer.GracefulStop()
	logrus.Infof("Server at %s is stopped", s.serverAddress)

	return nil
}

func (s *Server) startGRPC() {
	grpcListener := s.serverMux.Match(cmux.HTTP2())

	logrus.Infof("Starting server at %s", s.serverAddress)

	err := s.grpcServer.Serve(grpcListener)
	if err != nil {
		logrus.Errorf("error starting grpc server: %s", err)
	}
}

func (s *Server) startGateway() {
	httpListener := s.serverMux.Match(cmux.HTTP1Fast())

	httpMux := runtime.NewServeMux(
		runtime.WithMarshalerOption(
			runtime.MIMEWildcard, &runtime.JSONPb{}))

	err := velez_api.RegisterVelezAPIHandlerFromEndpoint(
		context.Background(),
		httpMux,
		s.serverAddress,
		[]grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		})
	if err != nil {
		logrus.Errorf("error registering grpc2http handler: %s", err)
	}
	server := &http.Server{
		Addr:    s.serverAddress,
		Handler: httpMux,
	}

	err = server.Serve(httpListener)
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			logrus.Errorf("error starting gateway: %s", err)
		}
	}
}
