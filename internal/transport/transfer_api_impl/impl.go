// Package transfer_api_impl implements the TransferAPI transport - decode, call
// service.TransferInService / service.TransferOutService, encode. No business
// logic.
package transfer_api_impl

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
	velez_api.UnimplementedTransferAPIServer

	transfersIn  service.TransferInService
	transfersOut service.TransferOutService
}

func New(srv service.Services) *Impl {
	return &Impl{
		UnimplementedTransferAPIServer: velez_api.UnimplementedTransferAPIServer{},

		transfersIn:  srv.TransfersIn(),
		transfersOut: srv.TransfersOut(),
	}
}

func (impl *Impl) Register(server grpc.ServiceRegistrar) {
	velez_api.RegisterTransferAPIServer(server, impl)
}

func (impl *Impl) Gateway(
	ctx context.Context,
	endpoint string,
	opts ...grpc.DialOption,
) (route string, handler http.Handler) {
	gwHTTPMux := runtime.NewServeMux()

	err := velez_api.RegisterTransferAPIHandlerFromEndpoint(
		ctx,
		gwHTTPMux,
		endpoint,
		opts,
	)
	if err != nil {
		log.Error().Err(err).Msg("error registering grpc2http handler")
	}

	return "/api/transfer/", gwHTTPMux
}
