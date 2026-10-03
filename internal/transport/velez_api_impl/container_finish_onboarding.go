package velez_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) FinishOnboarding(
	ctx context.Context,
	req *velez_api.FinishOnboarding_Request,
) (*velez_api.FinishOnboarding_Response, error) {
	_, err := impl.resolveEnvironment(ctx, req.GetEnvironment())
	if err != nil {
		return nil, err
	}

	resp, err := impl.smerdService.FinishOnboarding(ctx, req)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	return resp, nil
}
