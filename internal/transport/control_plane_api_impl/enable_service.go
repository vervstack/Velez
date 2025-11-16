package control_plane_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/pkg/velez_api"
)

func (impl *Impl) EnableService(ctx context.Context, req *velez_api.EnableServices_Request) (*velez_api.EnableServices_Response, error) {
	runner := impl.pipeliner.EnableVervService(req.Services[0])
	err := runner.Run(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	return &velez_api.EnableServices_Response{}, nil
}
