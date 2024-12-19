package container_service_task

import (
	errors "go.redsock.ru/rerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ApiClient[T any] struct {
	Client T
}

func NewGrpcClient[T any](addr string, constructor func(conn grpc.ClientConnInterface) T, opts ...grpc.DialOption) (*ApiClient[T], error) {
	opts = append(opts,
		grpc.WithTransportCredentials(insecure.NewCredentials()))

	dial, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "error dialing")
	}

	client := &ApiClient[T]{
		Client: constructor(dial),
	}

	return client, nil
}
