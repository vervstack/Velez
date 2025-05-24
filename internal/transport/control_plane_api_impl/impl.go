package control_plane_api_impl

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"go.vervstack.ru/Velez/internal/backservice/service_discovery"
	"go.vervstack.ru/Velez/pkg/control_plane_api"
)

type Impl struct {
	control_plane_api.UnimplementedControlPlaneServer

	sd service_discovery.ServiceDiscovery
}

func New(sd service_discovery.ServiceDiscovery) *Impl {
	return &Impl{
		sd: sd,
	}
}

func (impl *Impl) Register(server grpc.ServiceRegistrar) {
	control_plane_api.RegisterControlPlaneServer(server, impl)
}

func (impl *Impl) Gateway(ctx context.Context, endpoint string, opts ...grpc.DialOption) (route string, handler http.Handler) {
	gwHttpMux := runtime.NewServeMux()

	err := control_plane_api.RegisterControlPlaneHandlerFromEndpoint(
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
