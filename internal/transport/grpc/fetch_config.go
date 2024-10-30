package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/godverv/Velez/pkg/velez_api"
)

func (a *Impl) FetchConfig(ctx context.Context, req *velez_api.FetchConfig_Request) (*velez_api.FetchConfig_Response, error) {
	cfg, err := a.srv.FetchConfig(ctx, req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	resp := &velez_api.FetchConfig_Response{}

	resp.Config, err = cfg.Marshal()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}
