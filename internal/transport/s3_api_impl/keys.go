package s3_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

func (impl *Impl) ListS3Keys(
	ctx context.Context,
	req *pb.ListS3Keys_Request,
) (*pb.ListS3Keys_Response, error) {
	keys, err := impl.s3Service.ListKeys(ctx, req.GetInstanceName())
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing s3 keys")
	}

	out := make([]*pb.S3Key, 0, len(keys))
	for _, key := range keys {
		out = append(out, s3KeyToPb(key))
	}

	resp := &pb.ListS3Keys_Response{Keys: out}

	return resp, nil
}

func (impl *Impl) CreateS3Key(
	ctx context.Context,
	req *pb.CreateS3Key_Request,
) (*pb.CreateS3Key_Response, error) {
	access := make([]domain.S3BucketAccess, 0, len(req.GetAccess()))
	for _, item := range req.GetAccess() {
		access = append(access, s3BucketAccessFromPb(item))
	}

	credentials, err := impl.s3Service.CreateKey(ctx, req.GetInstanceName(), req.GetKeyName(), access)
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating s3 key")
	}

	resp := &pb.CreateS3Key_Response{
		AccessKeyId:     credentials.AccessKeyId,
		SecretAccessKey: credentials.SecretAccessKey,
	}

	return resp, nil
}

func (impl *Impl) DeleteS3Key(
	ctx context.Context,
	req *pb.DeleteS3Key_Request,
) (*pb.DeleteS3Key_Response, error) {
	err := impl.s3Service.DeleteKey(ctx, req.GetInstanceName(), req.GetAccessKeyId())
	if err != nil {
		return nil, rerrors.Wrap(err, "error deleting s3 key")
	}

	return &pb.DeleteS3Key_Response{}, nil
}

func (impl *Impl) GetS3KeyCredentials(
	ctx context.Context,
	req *pb.GetS3KeyCredentials_Request,
) (*pb.GetS3KeyCredentials_Response, error) {
	credentials, err := impl.s3Service.GetKeyCredentials(ctx, req.GetInstanceName(), req.GetAccessKeyId())
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting s3 key credentials")
	}

	resp := &pb.GetS3KeyCredentials_Response{
		AccessKeyId:        credentials.AccessKeyId,
		SecretAccessKey:    credentials.SecretAccessKey,
		S3Endpoint:         credentials.S3Endpoint,
		InternalS3Endpoint: credentials.InternalS3Endpoint,
		Region:             credentials.Region,
	}

	return resp, nil
}
