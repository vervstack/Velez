package service_discovery

import (
	"context"

	makosh "github.com/godverv/makosh/pkg/makosh_be"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func (s *ServiceDiscovery) Version(ctx context.Context, in *makosh.Version_Request, opts ...grpc.CallOption) (*makosh.Version_Response, error) {
	ctx = metadata.NewOutgoingContext(ctx, s.md)
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
