package s3_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/jobs"
)

func (impl *Impl) CreateS3Instance(
	ctx context.Context,
	req *pb.CreateS3Instance_Request,
) (*pb.CreateS3Instance_Response, error) {
	err := impl.s3Service.CreateInstance(ctx, req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating s3 instance")
	}

	resp := &pb.CreateS3Instance_Response{
		EntityId: req.GetName(),
		Action:   jobs.CreateS3InstanceAction,
	}

	return resp, nil
}
