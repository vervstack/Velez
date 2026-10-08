package runners_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) SetRunnerBuildkit(
	ctx context.Context,
	req *pb.SetRunnerBuildkit_Request,
) (*pb.SetRunnerBuildkit_Response, error) {
	taskId, err := impl.runnersService.SetRunnerBuildkit(ctx, req.GetName(), req.GetIsBuildkitEnabled())
	if err != nil {
		return nil, rerrors.Wrap(err, "error setting runner buildkit")
	}

	resp := &pb.SetRunnerBuildkit_Response{
		TaskId: taskId,
	}

	return resp, nil
}
