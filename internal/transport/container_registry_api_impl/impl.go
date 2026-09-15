// Package container_registry_api_impl implements the ContainerRegistryAPI
// transport - decode, call service.ContainerRegistryService, encode. No
// business logic - see docs/features/pgaas_and_registry_plugin.md section 4.
package container_registry_api_impl

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/service"
)

type Impl struct {
	velez_api.UnimplementedContainerRegistryAPIServer

	containerRegistryService service.ContainerRegistryService
}

func New(srv service.Services) *Impl {
	return &Impl{
		UnimplementedContainerRegistryAPIServer: velez_api.UnimplementedContainerRegistryAPIServer{},

		containerRegistryService: srv.ContainerRegistry(),
	}
}

func (impl *Impl) Register(server grpc.ServiceRegistrar) {
	velez_api.RegisterContainerRegistryAPIServer(server, impl)
}

func (impl *Impl) Gateway(
	ctx context.Context,
	endpoint string,
	opts ...grpc.DialOption,
) (route string, handler http.Handler) {
	gwHTTPMux := runtime.NewServeMux()

	err := velez_api.RegisterContainerRegistryAPIHandlerFromEndpoint(
		ctx,
		gwHTTPMux,
		endpoint,
		opts,
	)
	if err != nil {
		log.Error().Err(err).Msg("error registering grpc2http handler")
	}

	return "/api/container-registry/", gwHTTPMux
}
