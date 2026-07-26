package control_plane_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"
	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) ListPlugins(ctx context.Context, _ *pb.ListPlugins_Request) (*pb.ListPlugins_Response, error) {
	resp, err := impl.pluginService.ListPlugins(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing plugins")
	}

	return resp, nil
}
