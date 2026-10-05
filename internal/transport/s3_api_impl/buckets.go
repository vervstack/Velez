package s3_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) ListS3Buckets(
	ctx context.Context,
	req *pb.ListS3Buckets_Request,
) (*pb.ListS3Buckets_Response, error) {
	buckets, err := impl.s3Service.ListBuckets(ctx, req.GetInstanceName())
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing s3 buckets")
	}

	out := make([]*pb.S3Bucket, 0, len(buckets))
	for _, bucket := range buckets {
		out = append(out, s3BucketToPb(bucket))
	}

	resp := &pb.ListS3Buckets_Response{Buckets: out}

	return resp, nil
}

func (impl *Impl) CreateS3Bucket(
	ctx context.Context,
	req *pb.CreateS3Bucket_Request,
) (*pb.CreateS3Bucket_Response, error) {
	bucket, err := impl.s3Service.CreateBucket(ctx, req.GetInstanceName(), req.GetBucketName(), req.GetOwnerService())
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating s3 bucket")
	}

	resp := &pb.CreateS3Bucket_Response{Bucket: s3BucketToPb(bucket)}

	return resp, nil
}

func (impl *Impl) DeleteS3Bucket(
	ctx context.Context,
	req *pb.DeleteS3Bucket_Request,
) (*pb.DeleteS3Bucket_Response, error) {
	err := impl.s3Service.DeleteBucket(ctx, req.GetInstanceName(), req.GetBucketName())
	if err != nil {
		return nil, rerrors.Wrap(err, "error deleting s3 bucket")
	}

	return &pb.DeleteS3Bucket_Response{}, nil
}

func (impl *Impl) SetS3BucketAccess(
	ctx context.Context,
	req *pb.SetS3BucketAccess_Request,
) (*pb.SetS3BucketAccess_Response, error) {
	access := s3BucketAccessFromPb(req.GetAccess())

	bucket, err := impl.s3Service.SetBucketAccess(ctx, req.GetInstanceName(), access)
	if err != nil {
		return nil, rerrors.Wrap(err, "error setting s3 bucket access")
	}

	resp := &pb.SetS3BucketAccess_Response{Bucket: s3BucketToPb(bucket)}

	return resp, nil
}
