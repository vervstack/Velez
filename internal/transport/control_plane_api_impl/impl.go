package control_plane_api_impl

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
	velez_api.UnimplementedControlPlaneAPIServer

	pipeliner  pipelines.Pipeliner
	jobsEngine jobs.Engine

	smerdManager  service.ContainerService
	nodeService   service.NodeService
	pluginService service.PluginService
	vervServices  service.VervServicesService
}

func New(srv service.Services, pipeliner pipelines.Pipeliner, jobsEngine jobs.Engine) *Impl {
	return &Impl{
		UnimplementedControlPlaneAPIServer: velez_api.UnimplementedControlPlaneAPIServer{},
		pipeliner:                          pipeliner,
		jobsEngine:                         jobsEngine,

		smerdManager:  srv.SmerdManager(),
		nodeService:   srv.NodeService(),
		pluginService: srv.PluginService(),
		vervServices:  srv.VervServices(),
	}
}

func (impl *Impl) Register(server grpc.ServiceRegistrar) {
	velez_api.RegisterControlPlaneAPIServer(server, impl)
}

func (impl *Impl) Gateway(ctx context.Context, endpoint string, opts ...grpc.DialOption) (route string, handler http.Handler) {
	gwHttpMux := runtime.NewServeMux()

	err := velez_api.RegisterControlPlaneAPIHandlerFromEndpoint(
		ctx,
		gwHttpMux,
		endpoint,
		opts,
	)
	if err != nil {
		log.Error().Err(err).Msg("error registering grpc2http handler")
	}

	return "/api/control_plane/", gwHttpMux
}
