package velez_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

func (a *Impl) UpgradeSmerd(ctx context.Context,
	req *velez_api.UpgradeSmerd_Request) (*velez_api.UpgradeSmerd_Response, error) {

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
