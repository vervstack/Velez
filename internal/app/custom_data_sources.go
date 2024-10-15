package app

import (
	errors "github.com/Red-Sock/trace-errors"
	grpc2 "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/godverv/Velez/internal/clients/grpc"
)

func (a *App) InitDataSources() (err error) {
	a.GrpcMatreshkaBe, err = grpc.NewMatreshkaBeAPIClient(a.Cfg.DataSources.GrpcMatreshkaBe,
		grpc2.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return errors.Wrap(err, "error during grpc client initialization")
	}

	a.GrpcMakosh, err = grpc.NewMakoshBeAPIClient(a.Cfg.DataSources.GrpcMakosh,
		grpc2.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return errors.Wrap(err, "error during grpc client initialization")
	}

	return nil
}
