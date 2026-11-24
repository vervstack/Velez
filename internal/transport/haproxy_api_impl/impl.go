package haproxy_api_impl

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
	velez_api.UnimplementedHaproxyAPIServer

	haproxyService service.HaproxyService
	provisioning   service.ProvisioningService
}

func New(srv service.Services) *Impl {
	return &Impl{
		UnimplementedHaproxyAPIServer: velez_api.UnimplementedHaproxyAPIServer{},

		haproxyService: srv.Haproxy(),
		provisioning:   srv.Provisioning(),
	}
}

func (impl *Impl) Register(server grpc.ServiceRegistrar) {
	velez_api.RegisterHaproxyAPIServer(server, impl)
}

func (impl *Impl) Gateway(
	ctx context.Context,
	endpoint string,
	opts ...grpc.DialOption,
) (route string, handler http.Handler) {
	mux := runtime.NewServeMux()

	err := velez_api.RegisterHaproxyAPIHandlerFromEndpoint(
		ctx,
		mux,
		endpoint,
		opts,
	)
	if err != nil {
		log.Error().Err(err).Msg("error registering grpc2http handler")
	}

	return "/api/haproxy/", mux
}
