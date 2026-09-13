package runners_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

func (impl *Impl) CreateRunner(
	ctx context.Context,
	req *pb.CreateRunner_Request,
) (*pb.CreateRunner_Response, error) {
	serviceReq := domain.CreateRunnerReq{
		Name:        req.GetName(),
		Provider:    req.GetProvider(),
		Scope:       req.GetScope(),
		Target:      req.GetTarget(),
		Labels:      req.GetLabels(),
		AccessToken: req.GetAccessToken(),
		Environment: req.GetEnvironment(),
	}

	view, err := impl.runnersService.CreateRunner(ctx, serviceReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating runner")
	}

	resp := &pb.CreateRunner_Response{
		Runner: runnerToPb(view),
	}

	return resp, nil
}
