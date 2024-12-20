package control_plane_api_impl

import (
	"context"

	"github.com/godverv/Velez/pkg/control_plane_api"
)

func (impl *Impl) ListServices(ctx context.Context, request *control_plane_api.ListServices_Request) (
	*control_plane_api.ListServices_Response, error) {

	return &control_plane_api.ListServices_Response{
		Matreshka: &control_plane_api.Matreshka{},
		Makosh:    &control_plane_api.Makosh{},
	}, nil
}
