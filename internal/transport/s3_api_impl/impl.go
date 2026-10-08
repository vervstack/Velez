// Package s3_api_impl implements the S3API transport - decode, call
// service.S3Service, encode. No business logic.
package s3_api_impl

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
	velez_api.UnimplementedS3APIServer

	s3Service    service.S3Service
	provisioning service.ProvisioningService
}

func New(srv service.Services) *Impl {
	return &Impl{
		UnimplementedS3APIServer: velez_api.UnimplementedS3APIServer{},

		s3Service:    srv.S3(),
		provisioning: srv.Provisioning(),
	}
}

func (impl *Impl) Register(server grpc.ServiceRegistrar) {
	velez_api.RegisterS3APIServer(server, impl)
}

func (impl *Impl) Gateway(
	ctx context.Context,
	endpoint string,
	opts ...grpc.DialOption,
) (route string, handler http.Handler) {
	gwHTTPMux := runtime.NewServeMux()

	err := velez_api.RegisterS3APIHandlerFromEndpoint(
		ctx,
		gwHTTPMux,
		endpoint,
		opts,
	)
	if err != nil {
		log.Error().Err(err).Msg("error registering grpc2http handler")
	}

	return "/api/s3/", gwHTTPMux
}
