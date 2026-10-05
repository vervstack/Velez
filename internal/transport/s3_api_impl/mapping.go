package s3_api_impl

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

func s3InstanceToPb(instance domain.S3Instance) *pb.S3Instance {
	out := &pb.S3Instance{
		Name:              instance.Name,
		S3Port:            instance.S3Port,
		ReplicationFactor: instance.ReplicationFactor,
		Region:            instance.Region,
		Environment:       instance.Environment,
		Status:            instance.Status,
	}

	if instance.WebUiPort != 0 {
		out.WebUiPort = &instance.WebUiPort
	}

	if !instance.CreatedAt.IsZero() {
		out.CreatedAt = timestamppb.New(instance.CreatedAt)
	}

	return out
}

func s3BucketToPb(bucket domain.S3Bucket) *pb.S3Bucket {
	out := &pb.S3Bucket{
		Id:          bucket.Id,
		Name:        bucket.Name,
		Access:      s3BucketAccessListToPb(bucket.Access),
		ObjectCount: bucket.ObjectCount,
		Bytes:       bucket.Bytes,
	}

	if bucket.OwnerService != "" {
		out.OwnerService = &bucket.OwnerService
	}

	if !bucket.CreatedAt.IsZero() {
		out.CreatedAt = timestamppb.New(bucket.CreatedAt)
	}

	return out
}

func s3KeyToPb(key domain.S3Key) *pb.S3Key {
	out := &pb.S3Key{
		AccessKeyId: key.AccessKeyId,
		Name:        key.Name,
		Access:      s3BucketAccessListToPb(key.Access),
	}

	if key.OwnerService != "" {
		out.OwnerService = &key.OwnerService
	}

	if !key.CreatedAt.IsZero() {
		out.CreatedAt = timestamppb.New(key.CreatedAt)
	}

	return out
}

func s3BucketAccessListToPb(access []domain.S3BucketAccess) []*pb.S3BucketAccess {
	out := make([]*pb.S3BucketAccess, 0, len(access))
	for _, item := range access {
		out = append(out, s3BucketAccessToPb(item))
	}

	return out
}

func s3BucketAccessToPb(access domain.S3BucketAccess) *pb.S3BucketAccess {
	return &pb.S3BucketAccess{
		BucketName:  access.BucketName,
		AccessKeyId: access.AccessKeyId,
		KeyName:     access.KeyName,
		IsRead:      access.IsRead,
		IsWrite:     access.IsWrite,
		IsOwner:     access.IsOwner,
	}
}

func s3BucketAccessFromPb(access *pb.S3BucketAccess) domain.S3BucketAccess {
	return domain.S3BucketAccess{
		BucketName:  access.GetBucketName(),
		AccessKeyId: access.GetAccessKeyId(),
		KeyName:     access.GetKeyName(),
		IsRead:      access.GetIsRead(),
		IsWrite:     access.GetIsWrite(),
		IsOwner:     access.GetIsOwner(),
	}
}
