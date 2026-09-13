package runners_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/user_errors"
)

func (impl *Impl) CreateRunner(
	ctx context.Context,
	req *pb.CreateRunner_Request,
) (*pb.CreateRunner_Response, error) {
	serviceReq := domain.CreateRunnerReq{
		Name:                req.GetName(),
		Scope:               req.GetScope(),
		Target:              req.GetTarget(),
		Labels:              req.GetLabels(),
		Environment:         req.GetEnvironment(),
		DockerSocketAddress: req.GetDockerSocketAddress(),
	}

	switch cfg := req.GetProviderConfig().(type) {
	case *pb.CreateRunner_Request_Github:
		serviceReq.Provider = pb.RunnerProvider_GITHUB
		serviceReq.AccessToken = cfg.Github.GetAccessToken()
	default:
		return nil, rerrors.Wrap(user_errors.ErrRunnerProviderUnsupported)
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
