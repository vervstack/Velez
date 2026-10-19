package user_errors

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

var (
	// ErrAutoUpdateServiceNameRequired is returned by SetServiceAutoUpdate when
	// the request carries no service name.
	ErrAutoUpdateServiceNameRequired = rerrors.New("service name is required", codes.InvalidArgument)

	// ErrAutoUpdateServiceNotFound is returned by SetServiceAutoUpdate when the
	// name doesn't resolve to a plain app service of the default environment.
	ErrAutoUpdateServiceNotFound = rerrors.New("service not found", codes.NotFound)
)
