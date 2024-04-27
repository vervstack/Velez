package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/godverv/Velez/pkg/velez_api"
)

func (a *Api) ListSmerds(ctx context.Context, req *velez_api.ListSmerds_Request) (*velez_api.ListSmerds_Response, error) {
	err := req.ValidateAll()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return a.containerManager.ListSmerds(ctx, req)
}
