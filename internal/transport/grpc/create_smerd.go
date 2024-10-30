package grpc

import (
	"context"

	errors "github.com/Red-Sock/trace-errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/godverv/Velez/pkg/velez_api"
)

func (a *Impl) CreateSmerd(ctx context.Context, req *velez_api.CreateSmerd_Request) (*velez_api.Smerd, error) {
	err := req.ValidateAll()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(err, "invalid request. Must match \"lowercase/lowercase:v0.0.1\"").Error())
	}

	id, err := a.srv.LaunchSmerd(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "error launching smerd")
	}

	smerd, err := a.srv.InspectSmerd(ctx, id)
	if err != nil {
		return nil, err
	}

	return smerd, nil
}
