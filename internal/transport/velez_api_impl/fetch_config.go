package velez_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"

	"github.com/godverv/Velez/pkg/velez_api"
)

func (a *Impl) FetchConfig(ctx context.Context, req *velez_api.FetchConfig_Request) (*velez_api.FetchConfig_Response, error) {
	executor := a.pipeliner.GetConfig(req)
	err := executor.Run(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error during GetConfig pipeline execution")
	}

	cfg, err := executor.Result()
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting results")
	}
	if cfg == nil {
		return nil, rerrors.New("No config found", codes.NotFound)
	}

	err = a.srv.UpdateConfig(ctx, req.ServiceName, *cfg)
	if err != nil {
		return nil, rerrors.Wrap(err, "error updating config")
	}

	resp := &velez_api.FetchConfig_Response{}

	resp.Config, err = cfg.Marshal()
	if err != nil {
		return nil, rerrors.Wrap(err, codes.Internal)
	}

	return resp, nil
}
