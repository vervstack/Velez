package grpc

import (
	errors "github.com/Red-Sock/trace-errors"
	pb "github.com/godverv/makosh/pkg/makosh_be"
	"github.com/godverv/matreshka/resources"
	"google.golang.org/grpc"
)

type MakoshBeAPIClient pb.MakoshBeAPIClient

func NewMakoshBeAPIClient(grpcConn *resources.GRPC, opts ...grpc.DialOption) (MakoshBeAPIClient, error) {
	conn, err := connect(grpcConn.ConnectionString, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "error crating grpc client")
	}

	return pb.NewMakoshBeAPIClient(conn), nil
}
