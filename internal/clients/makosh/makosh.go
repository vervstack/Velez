package makosh

import (
	"context"

	errors "github.com/Red-Sock/trace-errors"
	makosh "github.com/godverv/makosh/pkg/makosh_be"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	pb "github.com/godverv/makosh/pkg/makosh_be"
)

const (
	MakoshAuthHeader = "Makosh-Auth"

	ServiceName = "makosh"
)

type ServiceDiscovery struct {
	makosh.MakoshBeAPIClient
}

func New(token string, opts ...grpc.DialOption) (*ServiceDiscovery, error) {
	opts = append(opts, grpc.WithUnaryInterceptor(HeaderInterceptor(token)))

	dial, err := grpc.NewClient("verv://"+ServiceName, opts...)
	//dial, err := grpc.NewClient("0.0.0.0:50051", opts...)
	if err != nil {
		return nil, errors.Wrap(err, "error dialing")
	}

	return &ServiceDiscovery{
		MakoshBeAPIClient: pb.NewMakoshBeAPIClient(dial),
	}, nil
}

func HeaderInterceptor(token string) func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	md := metadata.New(map[string]string{
		MakoshAuthHeader: token,
	})

	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = metadata.NewOutgoingContext(ctx, md)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
