package grpc_impl

import (
	"context"

	"github.com/godverv/Velez/pkg/velez_api"
)

func (a *Impl) ListSmerds(ctx context.Context, req *velez_api.ListSmerds_Request) (*velez_api.ListSmerds_Response, error) {
	return a.srv.ListSmerds(ctx, req)
}
