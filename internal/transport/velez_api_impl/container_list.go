package velez_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) ListContainers(
	ctx context.Context,
	req *velez_api.ListContainers_Request,
) (*velez_api.ListContainers_Response, error) {
	_, err := impl.resolveEnvironment(ctx, req.GetEnvironment())
	if err != nil {
		return nil, err
	}

	resp, err := impl.smerdService.ListContainers(ctx, req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing containers")
	}

	return resp, nil
}
