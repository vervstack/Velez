package user_errors

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

var (
	ErrHaproxyInstanceNotFound  = rerrors.New("HAProxy instance not found", codes.NotFound)
	ErrHaproxyInstanceExists    = rerrors.New("a service with this name already exists", codes.AlreadyExists)
	ErrHaproxyInvalidPort       = rerrors.New("HAProxy port must be between 1 and 65535", codes.InvalidArgument)
	ErrHaproxyInvalidMaster     = rerrors.New("a valid PostgreSQL instance name is required", codes.InvalidArgument)
	ErrHaproxyMasterEnvironment = rerrors.New(
		"the PostgreSQL instance must be deployed in the HAProxy environment", codes.FailedPrecondition)
	ErrHaproxyInstanceStopped = rerrors.New("the HAProxy instance must be running", codes.FailedPrecondition)
	ErrHaproxySelectionBusy   = rerrors.New("another HAProxy master selection is in progress", codes.Aborted)
	ErrHaproxyNotStable       = rerrors.NewUserError(
		"the HAProxy container did not stay running", codes.FailedPrecondition)
)
