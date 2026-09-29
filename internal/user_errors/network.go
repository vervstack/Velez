package user_errors

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

var (
	ErrNetworkNotFound = rerrors.NewUserError("network not found")
	ErrNoSuchContainer = rerrors.NewUserError("no such container")

	// ErrContainerFileNotFound is returned by ContainerRuntime.CopyFromContainer
	// when the requested path yields no regular file.
	ErrContainerFileNotFound = rerrors.New("no regular file at container path")

	ErrPortMustBeExposedForBinary = rerrors.NewUserError(
		"when running velez as a binary - the created container ports must be exposed",
		codes.FailedPrecondition,
	)
)
