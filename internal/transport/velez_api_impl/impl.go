package velez_api_impl

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"go.vervstack.ru/Velez/internal/clients"
	"go.vervstack.ru/Velez/internal/config"
	"go.vervstack.ru/Velez/internal/pipelines"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

type Impl struct {
	velez_api.UnimplementedVelezAPIServer

	version string

	hardwareManager clients.HardwareManager
	cfgService      service.ConfigurationService
	smerdService    service.ContainerService
	pipeliner       pipelines.Pipeliner
}

func NewImpl(cfg config.Config, srv service.Services, pipeliner pipelines.Pipeliner) *Impl {
	return &Impl{
		version:      cfg.AppInfo.Version,
		cfgService:   srv.ConfigurationService(),
		smerdService: srv.SmerdManager(),
		pipeliner:    pipeliner,
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
