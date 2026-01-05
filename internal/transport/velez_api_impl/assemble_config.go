package velez_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

func (impl *Impl) AssembleConfig(ctx context.Context, req *velez_api.AssembleConfig_Request) (*velez_api.AssembleConfig_Response, error) {
	pipeReq := domain.AssembleConfig{
		ServiceName: req.ServiceName,
		ImageName:   req.ImageName,
	}

	executor := impl.pipeliner.AssembleConfig(pipeReq)
	err := executor.Run(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error during AssembleConfig pipeline execution")
	}

	cfg, err := executor.Result()
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting results")
	}
	if cfg == nil {
		return nil, rerrors.New("No config found", codes.NotFound)
	}

	err = impl.cfgService.UpdateConfig(ctx, *cfg)
	if err != nil {
		if !rerrors.Is(err, cluster_clients.ErrServiceIsDisabled) {
			return nil, rerrors.Wrap(err, "error updating config")
		}
	}

	resp := velez_api.AssembleConfig_Response{
		Config: cfg.ContentRaw,
	}

	return &resp, nil
}
