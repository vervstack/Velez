package grpc

import (
	"context"

	"github.com/godverv/Velez/pkg/velez_api"
)

func (a *Api) Version(_ context.Context, _ *velez_api.Version_Request) (*velez_api.Version_Response, error) {
	return &velez_api.Version_Response{
		Version: a.version,
	}, nil
}
