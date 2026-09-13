// Package runners_api_impl implements the RunnersAPI transport - decode,
// call service.RunnersService, encode. No business logic.
package runners_api_impl

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
	velez_api.UnimplementedRunnersAPIServer

	runnersService service.RunnersService
}

func New(srv service.Services) *Impl {
	return &Impl{
		UnimplementedRunnersAPIServer: velez_api.UnimplementedRunnersAPIServer{},

		runnersService: srv.Runners(),
	}
}

func (impl *Impl) Register(server grpc.ServiceRegistrar) {
	velez_api.RegisterRunnersAPIServer(server, impl)
}

func (impl *Impl) Gateway(
	ctx context.Context,
	endpoint string,
	opts ...grpc.DialOption,
) (route string, handler http.Handler) {
	gwHTTPMux := runtime.NewServeMux()

	err := velez_api.RegisterRunnersAPIHandlerFromEndpoint(
		ctx,
		gwHTTPMux,
		endpoint,
		opts,
	)
	if err != nil {
		log.Error().Err(err).Msg("error registering grpc2http handler")
	}

	return "/api/runners/", gwHTTPMux
}
