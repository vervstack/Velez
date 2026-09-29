package velez_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) GetContainer(
	ctx context.Context,
	req *velez_api.GetContainer_Request,
) (*velez_api.DockerContainer, error) {
	_, err := impl.resolveEnvironment(ctx, req.GetEnvironment())
	if err != nil {
		return nil, err
	}

	resp, err := impl.smerdService.GetContainer(ctx, req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting container")
	}

	return resp, nil
}
