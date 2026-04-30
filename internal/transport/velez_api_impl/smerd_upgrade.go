package velez_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

func (impl *Impl) UpgradeSmerd(ctx context.Context,
	req *velez_api.UpgradeSmerd_Request) (*velez_api.UpgradeSmerd_Response, error) {

	upgradeReq := domain.UpgradeSmerd{
		Name:  req.Name,
		Image: req.Image,
	}

	err := impl.pipeliner.UpgradeSmerd(upgradeReq).Run(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	return &velez_api.UpgradeSmerd_Response{}, nil
}
