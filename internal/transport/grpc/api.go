package grpc

import (
	"github.com/godverv/Velez/internal/clients"
	"github.com/godverv/Velez/internal/config"
	"github.com/godverv/Velez/internal/service"
	"github.com/godverv/Velez/pkg/velez_api"
)

type Api struct {
	velez_api.UnimplementedVelezAPIServer

	version string

	// TODO что-то впихнуть
	hardwareManager clients.HardwareManager

	srv service.Services
}

func NewApi(cfg config.Config, srv service.Services) *Api {
	return &Api{
		version: cfg.GetAppInfo().Version,
		srv:     srv,
	}
}
