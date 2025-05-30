package velez_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"

	"go.vervstack.ru/Velez/internal/backservice/env"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

func (a *Impl) UpgradeSmerd(ctx context.Context,
	req *velez_api.UpgradeSmerd_Request) (*velez_api.UpgradeSmerd_Response, error) {

	id := env.GetContainerId()
	if id != nil {
		smerd, err := a.smerdService.InspectSmerd(ctx, req.Name)
		if err != nil {
			return nil, rerrors.Wrap(err)
		}
		if smerd.Uuid == *id {
			return nil, rerrors.NewUserError("Can't perform self upgrade", codes.Unimplemented)
		}
	}

	upgradeReq := domain.UpgradeSmerd{
		Name:  req.Name,
		Image: req.Image,
	}

	err := a.pipeliner.UpgradeSmerd(upgradeReq).Run(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	return &velez_api.UpgradeSmerd_Response{}, nil
}
