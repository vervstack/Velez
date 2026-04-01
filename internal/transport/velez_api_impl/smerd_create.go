package velez_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	velez_api "go.vervstack.ru/Velez/internal/api/server/api/grpc"
	"go.vervstack.ru/Velez/internal/domain"
)

func (impl *Impl) CreateSmerd(ctx context.Context, req *velez_api.CreateSmerd_Request) (*velez_api.Smerd, error) {
	launchPipe := impl.pipeliner.LaunchSmerd(domain.LaunchSmerd{CreateSmerd_Request: req})
	err := launchPipe.Run(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	res, err := launchPipe.Result()
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting result")
	}

	smerd, err := impl.smerdService.InspectSmerd(ctx, res.ContainerId)
	if err != nil {
		return nil, err
	}

	return smerd, nil
}
