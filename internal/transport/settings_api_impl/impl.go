// Package settings_api_impl implements the SettingsAPI transport - decode,
// call service.SettingsService, encode. No business logic.
package settings_api_impl

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service"
)

type Impl struct {
	velez_api.UnimplementedSettingsAPIServer

	settingsService service.SettingsService
}

func New(srv service.Services) *Impl {
	return &Impl{
		UnimplementedSettingsAPIServer: velez_api.UnimplementedSettingsAPIServer{},

		settingsService: srv.Settings(),
	}
}

func (impl *Impl) Register(server grpc.ServiceRegistrar) {
	velez_api.RegisterSettingsAPIServer(server, impl)
}

func (impl *Impl) Gateway(
	ctx context.Context,
	endpoint string,
	opts ...grpc.DialOption,
) (route string, handler http.Handler) {
	gwHTTPMux := runtime.NewServeMux()

	err := velez_api.RegisterSettingsAPIHandlerFromEndpoint(
		ctx,
		gwHTTPMux,
		endpoint,
		opts,
	)
	if err != nil {
		log.Error().Err(err).Msg("error registering grpc2http handler")
	}

	return "/api/settings/", gwHTTPMux
}

func settingsToPb(settings domain.Settings) *velez_api.Settings {
	return &velez_api.Settings{
		IsSysboxEnabled:          settings.IsSysboxEnabled,
		IsSysboxWhitelistIgnored: settings.IsSysboxWhitelistIgnored,
	}
}
