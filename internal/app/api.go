package app

import (
	"github.com/sirupsen/logrus"

	"github.com/godverv/Velez/internal/config"
	"github.com/godverv/Velez/internal/transport"
	"github.com/godverv/Velez/internal/transport/grpc"
)

func (a *App) mustInitAPI() {
	a.ServerManager = transport.NewManager()

	grpcConf, err := a.Cfg.GetServers().GRPC(config.ServerGrpc)
	if err != nil {
		logrus.Fatalf("error getting grpc from config: %s", err)
	}

	a.GrpcApi, err = grpc.NewServer(
		a.Cfg,
		grpcConf,
		a.Services,
		a.SecurityManager,
	)
	if err != nil {
		logrus.Fatalf("error creating grpc server: %s", err)
	}

	a.ServerManager.AddServer(a.GrpcApi)
}
