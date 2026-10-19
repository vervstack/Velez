package user_errors

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

var (
	// ErrSysboxRuntimeUnavailable is returned by the settings and dinds
	// services when sysbox is requested but sysbox-runc is not registered in
	// the Docker daemon.
	ErrSysboxRuntimeUnavailable = rerrors.New(
		"sysbox-runc runtime is not registered in the Docker daemon", codes.FailedPrecondition)

	// ErrDindInUse is returned by dinds.DropDind while a runner still uses
	// the DinD.
	ErrDindInUse = rerrors.New("dind is used by a runner", codes.FailedPrecondition)

	// ErrDindNotFound is returned when a dind name doesn't resolve to a
	// velez.dind_instances row.
	ErrDindNotFound = rerrors.New("dind not found", codes.NotFound)

	// ErrDindNameInvalid is returned by dinds.CreateDind for a name that is
	// not a valid DNS-safe service name.
	ErrDindNameInvalid = rerrors.New(
		"dind name must consist of lowercase letters, digits and \"-\"", codes.InvalidArgument)

	// ErrDindNameTaken is returned by dinds.CreateDind when a service with
	// the name already exists.
	ErrDindNameTaken = rerrors.New("a service with this name already exists", codes.AlreadyExists)
)
