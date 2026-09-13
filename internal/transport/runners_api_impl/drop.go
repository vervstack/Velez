package runners_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) DropRunner(
	ctx context.Context,
	req *pb.DropRunner_Request,
) (*pb.DropRunner_Response, error) {
	err := impl.runnersService.DropRunner(ctx, req.GetName())
	if err != nil {
		return nil, rerrors.Wrap(err, "error dropping runner")
	}

	return &pb.DropRunner_Response{}, nil
}
