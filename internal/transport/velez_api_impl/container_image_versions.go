package velez_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) ListContainerImageVersions(
	ctx context.Context,
	req *velez_api.ListContainerImageVersions_Request,
) (*velez_api.ListContainerImageVersions_Response, error) {
	resp, err := impl.imageVersions.ListContainerImageVersions(ctx, req)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	return resp, nil
}
