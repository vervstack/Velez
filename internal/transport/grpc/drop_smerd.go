package grpc

import (
	"context"

	"github.com/godverv/Velez/pkg/velez_api"
)

func (a *Api) DropSmerd(ctx context.Context, req *velez_api.DropSmerd_Request) (*velez_api.DropSmerd_Response, error) {
	return a.srv.DropSmerds(ctx, req)
}
