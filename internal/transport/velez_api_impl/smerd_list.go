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
	resp, err := impl.smerdService.ListSmerds(ctx, req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing smerds")
	}

	return resp, nil
}
