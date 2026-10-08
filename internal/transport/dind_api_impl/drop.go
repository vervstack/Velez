package dind_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/jobs"
)

func (impl *Impl) DropDind(
	ctx context.Context,
	req *pb.DropDind_Request,
) (*pb.DropDind_Response, error) {
	err := impl.dindService.DropDind(ctx, req.GetName())
	if err != nil {
		return nil, rerrors.Wrap(err, "error dropping dind")
	}

	resp := &pb.DropDind_Response{EntityId: req.GetName(), Action: jobs.DropDindAction}

	return resp, nil
}
