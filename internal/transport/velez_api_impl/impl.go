package velez_api_impl

import (
	"context"
	"net/http"

	"github.com/docker/docker/client"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog/log"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/config"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/pipelines"
	"go.vervstack.ru/Velez/internal/service"
	"google.golang.org/grpc"
)

type Impl struct {
	velez_api.UnimplementedVelezAPIServer

	version string

	hardwareManager node_clients.HardwareManager
	cfgService      service.ConfigurationService
	smerdService    service.ContainerService
	pipeliner       pipelines.Pipeliner
	jobsEngine      jobs.Engine

	dockerAPI client.APIClient
}

func NewImpl(cfg config.Config, srv service.Services, pipeliner pipelines.Pipeliner, jobsEngine jobs.Engine) *Impl {
	return &Impl{
		version:      cfg.AppInfo.Version,
		cfgService:   srv.ConfigurationService(),
		smerdService: srv.SmerdManager(),
		pipeliner:    pipeliner,
		jobsEngine:   jobsEngine,

		dockerAPI: srv.Docker().Client(),
	}
}

func (impl *Impl) Register(srv grpc.ServiceRegistrar) {
	velez_api.RegisterVelezAPIServer(srv, impl)
}

func (impl *Impl) Gateway(
	ctx context.Context,
	endpoint string,
	opts ...grpc.DialOption,
) (route string, handler http.Handler) {
	gwHTTPMux := runtime.NewServeMux()

	err := velez_api.RegisterVelezAPIHandlerFromEndpoint(
		ctx,
		gwHTTPMux,
		endpoint,
		opts,
	)
	if err != nil {
		log.Error().Err(err).Msg("error registering grpc2http handler")
	}

	return "/api/", gwHTTPMux
}
