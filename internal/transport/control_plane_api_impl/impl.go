package control_plane_api_impl

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/pipelines"
	"go.vervstack.ru/Velez/internal/service"
)

type Impl struct {
	velez_api.UnimplementedControlPlaneAPIServer

	pipeliner pipelines.Pipeliner

	smerdManager service.ContainerService
	nodeService  service.NodeService
}

func New(srv service.Services, pipeliner pipelines.Pipeliner) *Impl {
	return &Impl{
		velez_api.UnimplementedControlPlaneAPIServer{},
		pipeliner,

		srv.SmerdManager(),
		srv.NodeService(),
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
		logrus.Errorf("error registering grpc2http handler: %s", err)
	}

	return "/api/control_plane/", gwHttpMux
}
