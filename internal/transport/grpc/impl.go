package grpc

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/godverv/Velez/internal/clients"
	"github.com/godverv/Velez/internal/config"
	"github.com/godverv/Velez/internal/service"
	"github.com/godverv/Velez/pkg/velez_api"
)

type Impl struct {
	velez_api.UnimplementedVelezAPIServer

	version string

	// TODO что-то впихнуть
	hardwareManager clients.HardwareManager

	srv service.Services
}

func NewImpl(cfg config.Config, srv service.Services) *Impl {
	return &Impl{
		version: cfg.AppInfo.Version,
		srv:     srv,
	}
}

func (a *Impl) Register(srv *grpc.Server) {
	velez_api.RegisterVelezAPIServer(srv, a)
}

func (a *Impl) Gateway(ctx context.Context) (route string, handler http.Handler) {
	gwHttpMux := runtime.NewServeMux(
		runtime.WithMarshalerOption(
			runtime.MIMEWildcard, &runtime.JSONPb{},
		),
	)

	err := velez_api.RegisterVelezAPIHandlerServer(ctx, gwHttpMux, a)
	if err != nil {
		logrus.Errorf("error registering grpc2http handler: %s", err)
	}

	return "/api/*", gwHttpMux
}
