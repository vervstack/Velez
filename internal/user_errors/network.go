package user_errors

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

var (
	ErrNetworkNotFound = rerrors.NewUserError("network not found", codes.NotFound)
	ErrNoSuchContainer = rerrors.NewUserError("no such container", codes.NotFound)

	ErrNetworkNameEmpty     = rerrors.NewUserError("network name is required", codes.InvalidArgument)
	ErrNetworkNotEmpty      = rerrors.NewUserError("network still has connected containers", codes.FailedPrecondition)
	ErrNetworkNotManaged    = rerrors.NewUserError("network is not managed by velez", codes.FailedPrecondition)
	ErrNetworkAlreadyExists = rerrors.NewUserError("network already exists", codes.AlreadyExists)

	// ErrContainerFileNotFound is returned by ContainerRuntime.CopyFromContainer
	// when the requested path yields no regular file.
	ErrContainerFileNotFound = rerrors.New("no regular file at container path")

	ErrPortMustBeExposedForBinary = rerrors.NewUserError(
		"when running velez as a binary - the created container ports must be exposed",
		codes.FailedPrecondition,
	)
)
