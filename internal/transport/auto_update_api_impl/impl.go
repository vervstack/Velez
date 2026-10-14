// Package auto_update_api_impl implements the AutoUpdateAPI transport - decode,
// call service.AutoUpdateService, encode. No business logic.
package auto_update_api_impl

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
	velez_api.UnimplementedAutoUpdateAPIServer

	autoUpdateService service.AutoUpdateService
}

func New(srv service.Services) *Impl {
	return &Impl{
		UnimplementedAutoUpdateAPIServer: velez_api.UnimplementedAutoUpdateAPIServer{},

		autoUpdateService: srv.AutoUpdate(),
	}
}

func (impl *Impl) Register(server grpc.ServiceRegistrar) {
	velez_api.RegisterAutoUpdateAPIServer(server, impl)
}

func (impl *Impl) Gateway(
	ctx context.Context,
	endpoint string,
	opts ...grpc.DialOption,
) (route string, handler http.Handler) {
	gwHTTPMux := runtime.NewServeMux()

	err := velez_api.RegisterAutoUpdateAPIHandlerFromEndpoint(
		ctx,
		gwHTTPMux,
		endpoint,
		opts,
	)
	if err != nil {
		log.Error().Err(err).Msg("error registering grpc2http handler")
	}

	return "/api/auto_update/", gwHTTPMux
}
