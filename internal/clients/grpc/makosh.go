package grpc

import (
	errors "github.com/Red-Sock/trace-errors"
	pb "github.com/godverv/makosh/pkg/makosh_be"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/godverv/Velez/internal/config"
)

func NewMakoshBeAPIClient(cfg config.Config, opts ...grpc.DialOption) (pb.MakoshBeAPIClient, error) {
	connCfg, err := cfg.GetDataSources().GRPC(config.ResourceGrpcMakosh)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't find key"+config.ResourceGrpcMakosh+" grpc connection in config")
	}

	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := connect(connCfg.ConnectionString, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "error connection to "+connCfg.Module)
	}

	return pb.NewMakoshBeAPIClient(conn), nil
}
