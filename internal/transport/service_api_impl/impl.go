package service_api_impl

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/pipelines"
	"go.vervstack.ru/Velez/internal/service"
)

type Impl struct {
	velez_api.UnimplementedServiceApiServer

	servicesService service.VervServicesService

	pipeliner  pipelines.Pipeliner
	jobsEngine jobs.Engine
}

func New(pipeliner pipelines.Pipeliner, services service.Services, jobsEngine jobs.Engine) *Impl {
	return &Impl{
		servicesService: services.VervServices(),
		pipeliner:       pipeliner,
		jobsEngine:      jobsEngine,
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
		log.Error().Err(err).Msg("error registering grpc2http handler")
	}

	return "/api/service/", gwHttpMux
}
