package grpc_impl

import (
	"context"

	"github.com/godverv/Velez/pkg/velez_api"
)

func (a *Impl) Version(_ context.Context, _ *velez_api.Version_Request) (*velez_api.Version_Response, error) {
	return &velez_api.Version_Response{
		Version: a.version,
	}, nil
}
