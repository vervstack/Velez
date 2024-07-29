package service_discoverer

import (
	"context"

	rtb "github.com/Red-Sock/toolbox"
	errors "github.com/Red-Sock/trace-errors"
	makosh "github.com/godverv/makosh/pkg/makosh_be"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	grpcClients "github.com/godverv/Velez/internal/clients/grpc"
	"github.com/godverv/Velez/internal/config"
)

const (
	MakoshAuthHeader = "Makosh-Auth"
)

type ServiceDiscovery struct {
	authToken string
	cl        makosh.MakoshBeAPIClient

	md metadata.MD
}

func New(ctx context.Context, cfg config.Config) (*ServiceDiscovery, error) {
	envVar := cfg.GetEnvironment()

	token := envVar.MakoshKey
	if rtb.IsEmpty(token) {
		keyBytes, err := rtb.Random(256)
		if err != nil {
			return nil, errors.Wrap(err, "error generating random makosh key")
		}

		token = string(keyBytes)
	}

	cl, err := grpcClients.NewMakoshBeAPIClient(ctx, cfg,
		grpc.WithUnaryInterceptor(interceptor(token)),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, errors.Wrap(err, "error creating makosh grpc client")
	}

	return &ServiceDiscovery{
		authToken: token,
		cl:        cl,
	}, nil
}

func (s *ServiceDiscovery) GetToken() string {
	return s.authToken
}

func (s *ServiceDiscovery) Version(ctx context.Context, in *makosh.Version_Request, opts ...grpc.CallOption) (*makosh.Version_Response, error) {
	return s.cl.Version(ctx, in, opts...)
}

func (s *ServiceDiscovery) ListEndpoints(ctx context.Context, in *makosh.ListEndpoints_Request, opts ...grpc.CallOption) (*makosh.ListEndpoints_Response, error) {
	ctx = metadata.NewOutgoingContext(ctx, s.md)
	return s.cl.ListEndpoints(ctx, in, opts...)
}

func (s *ServiceDiscovery) UpsertEndpoints(ctx context.Context, in *makosh.UpsertEndpoints_Request, opts ...grpc.CallOption) (*makosh.UpsertEndpoints_Response, error) {
	ctx = metadata.NewOutgoingContext(ctx, s.md)
	return s.cl.UpsertEndpoints(ctx, in, opts...)
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
