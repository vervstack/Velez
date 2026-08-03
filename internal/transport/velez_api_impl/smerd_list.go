package velez_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) ListSmerds(
	ctx context.Context,
	req *velez_api.ListSmerds_Request,
) (*velez_api.ListSmerds_Response, error) {
	// Environment is required and must resolve to a real environment;
	// container_manager.ListSmerds re-resolves it into the label filter.
	_, err := impl.resolveEnvironment(ctx, req.GetEnvironment())
	if err != nil {
		return nil, err
	}

	resp, err := impl.smerdService.ListSmerds(ctx, req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing smerds")
	}

	return resp, nil
}
