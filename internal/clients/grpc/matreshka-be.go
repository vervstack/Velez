package grpc

import (
	errors "github.com/Red-Sock/trace-errors"
	pb "github.com/godverv/matreshka-be/pkg/matreshka_api"
	"github.com/godverv/matreshka/resources"
	"google.golang.org/grpc"
)

type MatreshkaBeAPIClient pb.MatreshkaBeAPIClient

func NewMatreshkaBeAPIClient(grpcConn *resources.GRPC, opts ...grpc.DialOption) (MatreshkaBeAPIClient, error) {
	conn, err := connect(grpcConn.ConnectionString, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "error crating grpc client")
	}

	return pb.NewMatreshkaBeAPIClient(conn), nil
}
