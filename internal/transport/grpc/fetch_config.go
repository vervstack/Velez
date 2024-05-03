package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/godverv/Velez/pkg/velez_api"
)

func (a *Api) FetchConfig(ctx context.Context, req *velez_api.FetchConfig_Request) (*velez_api.FetchConfig_Response, error) {
	err := a.smerdService.FetchConfig(ctx, req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &velez_api.FetchConfig_Response{}, nil
}
