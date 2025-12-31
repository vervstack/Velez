package service_api_impl

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"go.vervstack.ru/Velez/internal/pipelines"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

type Impl struct {
	velez_api.UnimplementedServiceApiServer

	servicesService service.VervServicesService

	pipeliner pipelines.Pipeliner
}

func New(pipeliner pipelines.Pipeliner, services service.Services) *Impl {
	return &Impl{
		servicesService: services.VervServices(),
		pipeliner:       pipeliner,
	}
}

func (impl *Impl) Register(srv grpc.ServiceRegistrar) {
	velez_api.RegisterServiceApiServer(srv, impl)
}

func (impl *Impl) Gateway(ctx context.Context, endpoint string, opts ...grpc.DialOption) (route string, handler http.Handler) {
	gwHttpMux := runtime.NewServeMux()

	err := velez_api.RegisterServiceApiHandlerFromEndpoint(
		ctx,
		gwHttpMux,
		endpoint,
		opts,
	)
	if err != nil {
		logrus.Errorf("error registering grpc2http handler: %s", err)
	}

	return "/api/service/", gwHttpMux
}
