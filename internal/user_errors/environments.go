package user_errors

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

var (
	// ErrEnvironmentRequired is returned when a request that must be scoped
	// to an environment carries an empty environment name. proto3 has no
	// `required` keyword, so this is where "required" actually gets
	// enforced.
	ErrEnvironmentRequired = rerrors.New("environment is required", codes.InvalidArgument)

	// ErrEnvironmentNotFound is returned when the environment name a caller
	// passed doesn't resolve to a row in velez.environments.
	ErrEnvironmentNotFound = rerrors.New("environment not found", codes.NotFound)

	// ErrEnvironmentIdRequired is returned when UpdateEnvironment carries no
	// ID.
	ErrEnvironmentIdRequired = rerrors.New("environment id is required", codes.InvalidArgument)

	// ErrEnvironmentIdOrNameRequired is returned when DeleteEnvironment
	// carries neither an ID nor a name to resolve the target by.
	ErrEnvironmentIdOrNameRequired = rerrors.New("environment id or name is required", codes.InvalidArgument)

	// ErrEnvironmentsStorageUnavailable is returned by
	// internal/storage/environments.Resolve when no EnvironmentsStorage
	// backend is wired in (single-node mode, before statefull mode is
	// enabled).
	ErrEnvironmentsStorageUnavailable = rerrors.New("environments storage is not available")
)
