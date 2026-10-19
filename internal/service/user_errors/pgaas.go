package user_errors

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

// ErrPgSharedPoolIsolationNotSupported is returned by
// PostgresService.CreatePgInstance for any isolation mode other than
// separate_instance - see docs/features/pgaas_and_registry_plugin.md section
// 3, "Isolation". shared_pool is surfaced disabled in the UI and has no
// provisioning path yet.
var ErrPgSharedPoolIsolationNotSupported = rerrors.NewUserError(
	"shared_pool isolation is not supported yet - only separate_instance is provisionable",
	codes.InvalidArgument,
)
