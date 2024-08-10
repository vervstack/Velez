package grpc

import (
	errors "github.com/Red-Sock/trace-errors"
	"google.golang.org/grpc"
)

func connect(connStr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	dial, err := grpc.NewClient(connStr, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "error dialing")
	}

	return dial, nil
}
