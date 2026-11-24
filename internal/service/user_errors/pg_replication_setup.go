package user_errors

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

var (
	ErrPgReplicationAlreadyPending = rerrors.NewUserError(
		"The target already has a different replication setup in progress.", codes.AlreadyExists,
	)
	ErrPgReplicationSourceNotReady = rerrors.NewUserError(
		"Source is not ready for streaming replication. Run Enable replication on the source first.",
		codes.FailedPrecondition,
	)
	ErrPgReplicationSourceInRecovery = rerrors.NewUserError(
		"Select a primary PostgreSQL instance as the source.", codes.FailedPrecondition,
	)
	ErrPgReplicationTargetNotEmpty = rerrors.NewUserError(
		"Select a fresh PostgreSQL target with no user tables in its default database and no existing replica configuration.",
		codes.FailedPrecondition,
	)
	ErrPgReplicationVolume = rerrors.NewUserError(
		"Replication requires a dedicated PostgreSQL 18 data volume on the target.", codes.FailedPrecondition,
	)
	ErrPgReplicationNetwork = rerrors.NewUserError(
		"Source and target must share a Docker network with an IP subnet.", codes.FailedPrecondition,
	)
	ErrPgReplicationVersion = rerrors.NewUserError(
		"Physical replication requires PostgreSQL 18 on both instances.", codes.FailedPrecondition,
	)
	ErrPgReplicationTablespaces = rerrors.NewUserError(
		"Replication setup requires the source to use PostgreSQL's built-in tablespaces.", codes.FailedPrecondition,
	)
	ErrPgReplicationTargetRunning = rerrors.NewUserError(
		"The target container must be stopped before copying the source.", codes.FailedPrecondition,
	)
	ErrPgReplicationOperationMissing = rerrors.New("replication task is missing a valid operation or target container id")
	ErrPgReplicationSlotUnavailable  = rerrors.NewUserError(
		"The replication slot already exists and is active or belongs to a logical replica.", codes.FailedPrecondition,
	)
	ErrPgReplicationAuth = rerrors.NewUserError(
		"Could not configure the source to accept streaming replication connections.", codes.FailedPrecondition,
	)
	ErrPgReplicationBackupFailed = rerrors.NewUserError(
		"PostgreSQL base backup failed. Check the source logs for replication connection or backup errors.",
		codes.FailedPrecondition,
	)
	ErrPgReplicationNotStreaming = rerrors.NewUserError(
		"The target did not start streaming from the source within 60 seconds.", codes.DeadlineExceeded,
	)
	ErrPgReplicationUnsafeRole = rerrors.New("source role name cannot be represented in pg_hba.conf")
	ErrPgReplicationMarker     = rerrors.New("invalid PostgreSQL replica metadata in the target data directory")
)
