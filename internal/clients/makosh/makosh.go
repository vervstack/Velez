package makosh

import (
	"context"

	errors "github.com/Red-Sock/trace-errors"
	makosh "github.com/godverv/makosh/pkg/makosh_be"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	grpcClients "github.com/godverv/Velez/internal/clients/grpc"
	"github.com/godverv/Velez/internal/config"
)

const (
	MakoshAuthHeader = "Makosh-Auth"
)

type ServiceDiscovery struct {
	makosh.MakoshBeAPIClient
}

func New(cfg config.Config, token string, opts ...grpc.DialOption) (*ServiceDiscovery, error) {
	opts = append(opts, grpc.WithUnaryInterceptor(interceptor(token)))

	cl, err := grpcClients.NewMakoshBeAPIClient(cfg.DataSources.GrpcMakosh, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "error creating makosh grpc client")
	}

	return &ServiceDiscovery{
		MakoshBeAPIClient: cl,
	}, nil
}

func interceptor(token string) func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	md := metadata.New(map[string]string{
		MakoshAuthHeader: token,
	})

	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = metadata.NewOutgoingContext(ctx, md)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
