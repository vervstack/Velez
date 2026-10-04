// Package dind_api_impl implements the DindAPI transport - decode, call
// service.DindService, encode. No business logic.
package dind_api_impl

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
	velez_api.UnimplementedDindAPIServer

	dindService service.DindService
}

func New(srv service.Services) *Impl {
	return &Impl{
		UnimplementedDindAPIServer: velez_api.UnimplementedDindAPIServer{},

		dindService: srv.Dinds(),
	}
}

func (impl *Impl) Register(server grpc.ServiceRegistrar) {
	velez_api.RegisterDindAPIServer(server, impl)
}

func (impl *Impl) Gateway(
	ctx context.Context,
	endpoint string,
	opts ...grpc.DialOption,
) (route string, handler http.Handler) {
	gwHTTPMux := runtime.NewServeMux()

	err := velez_api.RegisterDindAPIHandlerFromEndpoint(
		ctx,
		gwHTTPMux,
		endpoint,
		opts,
	)
	if err != nil {
		log.Error().Err(err).Msg("error registering grpc2http handler")
	}

	return "/api/dind/", gwHTTPMux
}
