// Package pgaas_api_impl implements the PostgresAPI transport - decode,
// call service.PostgresService, encode. No business logic - see
// docs/features/pgaas_and_registry_plugin.md section 3.
package pgaas_api_impl

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
	velez_api.UnimplementedPostgresAPIServer

	postgresService service.PostgresService
}

func New(srv service.Services) *Impl {
	return &Impl{
		UnimplementedPostgresAPIServer: velez_api.UnimplementedPostgresAPIServer{},

		postgresService: srv.Postgres(),
	}
}

func (impl *Impl) Register(server grpc.ServiceRegistrar) {
	velez_api.RegisterPostgresAPIServer(server, impl)
}

func (impl *Impl) Gateway(
	ctx context.Context,
	endpoint string,
	opts ...grpc.DialOption,
) (route string, handler http.Handler) {
	gwHTTPMux := runtime.NewServeMux()

	err := velez_api.RegisterPostgresAPIHandlerFromEndpoint(
		ctx,
		gwHTTPMux,
		endpoint,
		opts,
	)
	if err != nil {
		log.Error().Err(err).Msg("error registering grpc2http handler")
	}

	return "/api/postgres/", gwHTTPMux
}
