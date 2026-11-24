package user_errors

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

var (
	ErrPgFailoverStrategy = rerrors.NewUserError(
		"Choose rewind or recopy for the old master.", codes.InvalidArgument,
	)
	ErrPgFailoverNoOverride = rerrors.NewUserError(
		"Failover requires a fully synchronized replica.", codes.InvalidArgument,
	)
	ErrPgFailoverBusy = rerrors.NewUserError(
		"The balancer already has a different failover in progress.", codes.AlreadyExists,
	)
	ErrPgFailoverUpstream = rerrors.NewUserError(
		"The new master must be a PostgreSQL replica of the balancer's current master.", codes.FailedPrecondition,
	)
	ErrPgFailoverOutOfSync = rerrors.NewUserError(
		"Replica is not fully in sync yet", codes.FailedPrecondition,
	)
	ErrPgFailoverTopologyChanged = rerrors.NewUserError(
		"The PostgreSQL or balancer containers changed during failover.", codes.FailedPrecondition,
	)
	ErrPgFailoverPromoteTimeout = rerrors.NewUserError(
		"The replica did not leave recovery within 60 seconds.", codes.DeadlineExceeded,
	)
	ErrPgFailoverRewindHints = rerrors.NewUserError(
		"pg_rewind requires wal_log_hints on the old master. Choose Wipe and re-copy.", codes.FailedPrecondition,
	)
	ErrPgFailoverRewindFailed = rerrors.NewUserError(
		"pg_rewind failed. Keep the old master stopped and restore it from a fresh base backup.", codes.FailedPrecondition,
	)
)
