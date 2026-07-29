package user_errors

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

var (
	ErrNetworkNotFound = rerrors.NewUserError("network not found")
	ErrNoSuchContainer = rerrors.NewUserError("no such container")

	ErrPortMustBeExposedForBinary = rerrors.NewUserError(
		"when running velez as a binary - the created container ports must be exposed",
		codes.FailedPrecondition,
	)
)
