package user_errors

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

var ErrRegistryS3InstanceNameRequired = rerrors.New(
	"s3 instance name is required for s3 registry storage", codes.InvalidArgument)
