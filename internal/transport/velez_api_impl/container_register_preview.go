package velez_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/service/user_errors"
)

func (impl *Impl) PreviewRegisterContainer(
	ctx context.Context,
	req *velez_api.PreviewRegisterContainer_Request,
) (*velez_api.PreviewRegisterContainer_Response, error) {
	if req.GetContainerId() == "" {
		return nil, rerrors.Wrap(user_errors.ErrRegisterContainerIdRequired)
	}

	if req.GetServiceName() == "" {
		return nil, rerrors.Wrap(user_errors.ErrRegisterServiceNameRequired)
	}

	_, err := impl.resolveEnvironment(ctx, req.GetEnvironment())
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	err = impl.rejectUnregistrable(ctx, req.GetEnvironment(), req.GetContainerId())
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	resp, err := impl.smerdService.PreviewRegisterContainer(ctx, req)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	return resp, nil
}
