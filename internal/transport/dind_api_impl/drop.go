package dind_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) DropDind(
	ctx context.Context,
	req *pb.DropDind_Request,
) (*pb.DropDind_Response, error) {
	err := impl.dindService.DropDind(ctx, req.GetName())
	if err != nil {
		return nil, rerrors.Wrap(err, "error dropping dind")
	}

	return &pb.DropDind_Response{}, nil
}
