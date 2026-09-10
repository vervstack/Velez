package user_errors

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

// ErrRequiresStatefullMode is returned by the in-memory registries and
// environments storages (single-node/dev mode, no postgres) for every write
// method - that mode offers a fixed, baked-in set instead of mutable CRUD.
var ErrRequiresStatefullMode = rerrors.NewUserError(
	"not supported in single-node mode; enable statefull/postgres mode for registry and environment management",
	codes.FailedPrecondition,
)
