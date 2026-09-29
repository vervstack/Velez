package runners_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

const (
	redeployRunnerAction = "redeploy_runner"
)

func (impl *Impl) RedeployRunner(
	ctx context.Context,
	req *pb.RedeployRunner_Request,
) (*pb.RedeployRunner_Response, error) {
	err := impl.runnersService.RedeployRunner(ctx, req.GetName())
	if err != nil {
		return nil, rerrors.Wrap(err, "error redeploying runner")
	}

	resp := &pb.RedeployRunner_Response{
		EntityId: req.GetName(),
		Action:   redeployRunnerAction,
	}

	return resp, nil
}
