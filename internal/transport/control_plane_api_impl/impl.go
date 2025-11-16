package control_plane_api_impl

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
	velez_api.UnimplementedControlPlaneAPIServer

	smerdManager service.ContainerService

	pipeliner pipelines.Pipeliner
}

func New(srv service.Services, pipeliner pipelines.Pipeliner) *Impl {
	return &Impl{
		smerdManager: srv.SmerdManager(),
		pipeliner:    pipeliner,
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
