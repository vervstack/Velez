package velez_api_impl

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/godverv/Velez/internal/clients"
	"github.com/godverv/Velez/internal/config"
	"github.com/godverv/Velez/internal/pipelines"
	"github.com/godverv/Velez/internal/service"
	"github.com/godverv/Velez/pkg/velez_api"
)

type Impl struct {
	velez_api.UnimplementedVelezAPIServer

	version string

	// TODO что-то впихнуть
	hardwareManager clients.HardwareManager

	srv       service.Services
	pipeliner pipelines.Pipeliner
}

func NewImpl(cfg config.Config, srv service.Services, pipeliner pipelines.Pipeliner) *Impl {
	return &Impl{
		version:   cfg.AppInfo.Version,
		srv:       srv,
		pipeliner: pipeliner,
	}
}

func (a *Impl) Register(srv grpc.ServiceRegistrar) {
	velez_api.RegisterVelezAPIServer(srv, a)
}

func (a *Impl) Gateway(ctx context.Context, endpoint string, opts ...grpc.DialOption) (route string, handler http.Handler) {
	gwHttpMux := runtime.NewServeMux()

	err := velez_api.RegisterVelezAPIHandlerFromEndpoint(
		ctx,
		gwHttpMux,
		endpoint,
		opts,
	)
	if err != nil {
		logrus.Errorf("error registering grpc2http handler: %s", err)
	}

	return "/api/", gwHttpMux
}
