package jobs

import (
	"context"
	"strconv"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/service/user_errors"
)

const (
	pgRewindHelperPrefix  = "velez-pg-rewind-"
	pgRewindHintsExitCode = 42
	pgRewindScript        = `
data="$1"
source="$2"
operation="$3"
marker="$4"
slot="$5"
hints_exit="$6"
log=$(mktemp)
trap 'rm -f -- "$log"' EXIT

if ! pg_rewind --target-pgdata="$data" --source-server='sslmode=disable' --write-recovery-conf >"$log" 2>&1; then
    if grep -Fq 'wal_log_hints = on' "$log"; then
        exit "$hints_exit"
    fi
    exit 1
fi

test -f "$data/standby.signal"
printf "\nprimary_slot_name = '%s'\n" "$slot" >> "$data/postgresql.auto.conf"
sync -f "$data"
umask 077
printf '%s\n%s\n' "$source" "$operation" > "$data/$marker.tmp"
mv -f -- "$data/$marker.tmp" "$data/$marker"
sync -f "$data"
`
)

func (s *pgReplicationSetup) rewind(ctx context.Context) error {
	topology, err := s.backupTopology(ctx)
	if err != nil {
		return rerrors.Wrap(err)
	}

	if topology.target.State == nil || topology.target.State.Running {
		return rerrors.Wrap(user_errors.ErrPgReplicationTargetRunning)
	}

	name := pgRewindHelperPrefix + s.payload.GetOperationId()
	helper, err := s.docker.Client().ContainerInspect(ctx, name)

	if errdefs.IsNotFound(err) {
		return s.startRewind(ctx, topology, name)
	}

	if err != nil {
		return rerrors.Wrap(err, "error looking up pg_rewind helper")
	}

	if helper.State == nil {
		return rerrors.Wrap(user_errors.ErrPgFailoverRewindFailed)
	}

	// A running helper survives a worker crash. An interrupted pg_rewind can
	// leave partially rewritten files, so only start a helper that never ran.
	return s.finishRewind(ctx, helper.ID, helper.State.Status == "created")
}

func (s *pgReplicationSetup) startRewind(ctx context.Context, topology pgBackupTopology, name string) error {
	sourceName, operationId, err := readPgReplicaOperation(
		ctx, s.docker.Client(), s.payload.GetTargetContainerId(),
	)
	if err != nil {
		return rerrors.Wrap(err)
	}

	if isPgRewindComplete(s.payload, sourceName, operationId) {
		return nil
	}

	source, err := s.resolve(ctx, s.payload.GetSourceInstance())
	if err != nil {
		return rerrors.Wrap(err)
	}

	config := &container.Config{
		Image:      replicationExecutorImage,
		User:       pgSourceProcessUser,
		Entrypoint: []string{"sh", pgBackupShellOptions, pgRewindScript, "sh"},
		Cmd: []string{
			pgReplicaDataPath, source.host, s.payload.GetOperationId(), pgReplicaMarkerFile, s.slotName(),
			strconv.Itoa(pgRewindHintsExitCode),
		},
		Env: append(s.helperConnectionEnv(source), "PGDATABASE="+source.dbName),
	}

	id, err := s.createReplicationHelper(ctx, topology, config, name)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return s.finishRewind(ctx, id, true)
}

func isPgRewindComplete(payload *velez_api.SetupPgReplicationTaskPayload, sourceName, operationId string) bool {
	return operationId != "" && operationId == payload.GetOperationId() && sourceName == payload.GetSourceInstance()
}

func (s *pgReplicationSetup) finishRewind(ctx context.Context, helperId string, shouldStart bool) error {
	status, err := s.waitForReplicationHelper(ctx, helperId, shouldStart)
	if err != nil {
		return rerrors.Wrap(err)
	}

	err = s.removeReplicationHelper(ctx, helperId)
	if err != nil {
		return rerrors.Wrap(err)
	}

	switch {
	case status.Error != nil:
		return rerrors.Wrap(user_errors.ErrPgFailoverRewindFailed)
	case status.StatusCode == pgRewindHintsExitCode:
		return rerrors.Wrap(user_errors.ErrPgFailoverRewindHints)
	case status.StatusCode != 0:
		return rerrors.Wrap(user_errors.ErrPgFailoverRewindFailed)
	default:
		return nil
	}
}
