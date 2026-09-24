package runners_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/jobs"
)

func (impl *Impl) ReregisterRunner(
	ctx context.Context,
	req *pb.ReregisterRunner_Request,
) (*pb.ReregisterRunner_Response, error) {
	err := impl.runnersService.ReregisterRunner(ctx, req.GetName())
	if err != nil {
		return nil, rerrors.Wrap(err, "error reregistering runner")
	}

	resp := &pb.ReregisterRunner_Response{
		EntityId: req.GetName(),
		Action:   jobs.ReregisterRunnerAction,
	}

	return resp, nil
}
