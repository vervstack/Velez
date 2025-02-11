package velez_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	"github.com/godverv/Velez/internal/domain"
	"github.com/godverv/Velez/pkg/velez_api"
)

func (a *Impl) CreateSmerd(ctx context.Context, req *velez_api.CreateSmerd_Request) (*velez_api.Smerd, error) {
	pipe := a.pipeliner.LaunchSmerd(domain.LaunchSmerd{CreateSmerd_Request: req})
	err := pipe.Run(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	res, err := pipe.Result()
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting result")
	}

	smerd, err := a.smerdService.InspectSmerd(ctx, res.ContainerId)
	if err != nil {
		return nil, err
	}

	return smerd, nil
}
