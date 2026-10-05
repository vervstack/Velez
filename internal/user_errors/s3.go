package user_errors

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

var (
	ErrS3InstanceNameRequired = rerrors.New("s3 instance name is required", codes.InvalidArgument)

	ErrS3InstanceNameInvalid = rerrors.New(
		"s3 instance name must consist of lowercase letters, digits and \"-\"", codes.InvalidArgument)

	ErrS3InstanceAlreadyExists = rerrors.New("a service with this name already exists", codes.AlreadyExists)

	ErrS3InstanceNotFound = rerrors.New("s3 instance not found", codes.NotFound)

	ErrS3ReplicationFactorUnsupported = rerrors.New(
		"only replication factor 1 is supported", codes.InvalidArgument)

	ErrS3InstanceInUse = rerrors.New(
		"s3 instance is used by another service", codes.FailedPrecondition)

	ErrS3BucketNameRequired = rerrors.New("s3 bucket name is required", codes.InvalidArgument)

	ErrS3BucketNotFound = rerrors.New("s3 bucket not found", codes.NotFound)

	ErrS3KeyNotFound = rerrors.New("s3 key not found", codes.NotFound)

	ErrS3BucketAlreadyExists = rerrors.New("s3 bucket already exists", codes.AlreadyExists)

	ErrS3BucketNotEmpty = rerrors.New("s3 bucket is not empty", codes.FailedPrecondition)

	ErrS3KeyNameRequired = rerrors.New("s3 key name is required", codes.InvalidArgument)
)
