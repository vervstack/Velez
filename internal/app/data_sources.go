package app

import (
	errors "github.com/Red-Sock/trace-errors"
	"github.com/godverv/Velez/internal/clients/grpc"
)

func (a *App) InitDataSources() (err error) {
	a.GrpcMatreshkaBe, err = grpc.NewMatreshkaBeAPIClient(a.Cfg.DataSources.GrpcMatreshkaBe)
	if err != nil {
		return errors.Wrap(err, "error during grpc client initialization")
	}

	a.GrpcMakosh, err = grpc.NewMakoshBeAPIClient(a.Cfg.DataSources.GrpcMakosh)
	if err != nil {
		return errors.Wrap(err, "error during grpc client initialization")
	}

	a.GrpcMatreshkaBe, err = grpc.NewMatreshkaBeAPIClient(a.Cfg.DataSources.GrpcMatreshkaBe)
	if err != nil {
		return errors.Wrap(err, "error during grpc client initialization")
	}

	a.GrpcMakosh, err = grpc.NewMakoshBeAPIClient(a.Cfg.DataSources.GrpcMakosh)
	if err != nil {
		return errors.Wrap(err, "error during grpc client initialization")
	}

	return nil
}
