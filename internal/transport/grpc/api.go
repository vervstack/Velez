package grpc

import (
	"github.com/godverv/Velez/internal/clients"
	"github.com/godverv/Velez/internal/config"
	"github.com/godverv/Velez/internal/service"
	"github.com/godverv/Velez/pkg/velez_api"
)

type Impl struct {
	velez_api.UnimplementedVelezAPIServer

	version string

	// TODO что-то впихнуть
	hardwareManager clients.HardwareManager

	srv service.Services
}

func NewImpl(cfg config.Config, srv service.Services) *Impl {
	return &Impl{
		version: cfg.AppInfo.Version,
		srv:     srv,
	}
}
