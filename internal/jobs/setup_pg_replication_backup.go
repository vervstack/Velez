package jobs

import (
	"bytes"
	"context"
	"database/sql"
	"io"
	"net"
	"path"
	"slices"
	"strconv"
	"strings"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/service/user_errors"
)

const (
	pgBackupRunAsRoot      = "root"
	pgSourceProcessUser    = "postgres"
	pgBackupShellOptions   = "-ec"
	pgVolumePath           = "/var/lib/postgresql"
	pgReplicaDataPath      = pgVolumePath + "/18/docker"
	pgReplicaMarkerFile    = ".velez-replica-source"
	pgBackupHelperPrefix   = "velez-pg-basebackup-"
	pgBackupKilledExitCode = 137

	// The original cluster stays intact until a complete, fsynced standby exists.
	// Keep the staging and original directories on the same volume for atomic renames.
	pgBasebackupScript = `
data="$1"
source="$2"
operation="$3"
marker="$4"
slot="$5"
stage="$data.velez-stage-$operation"
original="$data.velez-original-$operation"

if [ -f "$data/$marker" ] && [ "$(sed -n '2p' "$data/$marker")" = "$operation" ]; then
    rm -rf -- "$original" "$stage"
    exit 0
fi

if [ -d "$original" ] && [ ! -e "$data" ]; then
    mv -T -- "$original" "$data"
fi
test -s "$data/PG_VERSION"
test ! -e "$original"

restore_original() {
    if [ -d "$original" ] && [ ! -e "$data" ]; then
        mv -T -- "$original" "$data"
    fi
    rm -rf -- "$stage"
}
trap restore_original EXIT
trap 'exit 1' HUP INT TERM

# Verify network and replication authentication before reserving space for the backup.
gosu postgres psql -X -w -qAt -d 'replication=true dbname=replication' -c IDENTIFY_SYSTEM >/dev/null

rm -rf -- "$stage"
install -d -m 0700 -o postgres -g postgres -- "$stage"
set -- pg_basebackup -w -D "$stage" -X stream --compress=server-gzip -R --checkpoint=fast
if [ -n "$slot" ]; then
    set -- "$@" -S "$slot"
fi
gosu postgres "$@"

test -s "$stage/PG_VERSION"
test -f "$stage/standby.signal"
tablespaces=$(ls -A -- "$stage/pg_tblspc")
test -z "$tablespaces"
gosu postgres sh -c 'umask 077; printf "%s\n%s\n" "$2" "$3" > "$1"' sh "$stage/$marker" "$source" "$operation"
sync -f "$stage"
mv -T -- "$data" "$original"
mv -T -- "$stage" "$data"
sync -f "$data"
rm -rf -- "$original"
`
)

type pgBackupTopology struct {
	sourceId string
	target   container.InspectResponse
	volume   string
	networks []string
}

func (s *pgReplicationSetup) backupTopology(ctx context.Context) (pgBackupTopology, error) {
	sourceId, err := FindPgInstanceContainerId(ctx, s.docker, s.payload.GetSourceInstance())
	if err != nil {
		return pgBackupTopology{}, rerrors.Wrap(err)
	}

	targetId := s.payload.GetTargetContainerId()
	if sourceId == targetId {
		return pgBackupTopology{}, rerrors.Wrap(user_errors.ErrPgReplicationSameInstance)
	}

	cli := s.docker.Client()

	source, err := cli.ContainerInspect(ctx, sourceId)
	if err != nil {
		return pgBackupTopology{}, rerrors.Wrap(err, "error inspecting replication source")
	}

	target, err := cli.ContainerInspect(ctx, targetId)
	if err != nil {
		return pgBackupTopology{}, rerrors.Wrap(err, "error inspecting replication target")
	}

	volume := pgReplicationTargetVolume(target)
	if volume == "" || pgMountsVolume(source, volume) {
		return pgBackupTopology{}, rerrors.Wrap(user_errors.ErrPgReplicationVolume)
	}

	networks := pgSharedNetworks(source.NetworkSettings, target.NetworkSettings)
	if len(networks) == 0 {
		return pgBackupTopology{}, rerrors.Wrap(user_errors.ErrPgReplicationNetwork)
	}

	topology := pgBackupTopology{sourceId: sourceId, target: target, volume: volume, networks: networks}

	return topology, nil
}

func (s *pgReplicationSetup) ensureReplicationAuth(
	ctx context.Context, db *sql.DB, topology pgBackupTopology, role string,
) error {
	if strings.ContainsAny(role, "\r\n") {
		return rerrors.Wrap(user_errors.ErrPgReplicationUnsafeRole)
	}

	var hbaPath string

	err := db.QueryRowContext(ctx, "SHOW hba_file").Scan(&hbaPath)
	if err != nil {
		return rerrors.Wrap(err, "error reading source hba_file location")
	}

	cli := s.docker.Client()

	subnets, err := pgReplicationSubnets(ctx, cli, topology.networks)
	if err != nil {
		return rerrors.Wrap(err)
	}

	// Physical replication does not match the stock image's "host all all" rule.
	// Keep the additional password-authenticated rule scoped to the shared networks.
	script := `grep -Fqx -- "$2" "$1" || printf '\n%s\n' "$2" >> "$1"`

	for _, subnet := range subnets {
		rule := "host replication " + quotePgIdentifier(role) + " " + subnet + " scram-sha-256"
		opts := container.ExecOptions{
			User:         pgSourceProcessUser,
			Cmd:          []string{"sh", pgBackupShellOptions, script, "sh", hbaPath, rule},
			AttachStdout: true,
			AttachStderr: true,
		}

		err = runPgReplicationExec(ctx, cli, topology.sourceId, opts)
		if err != nil {
			return rerrors.Wrap(err, "error authorizing physical replication on the source")
		}
	}

	var isReloading bool

	err = db.QueryRowContext(ctx, "SELECT pg_reload_conf()").Scan(&isReloading)
	if err != nil {
		return rerrors.Wrap(err, "error reloading source authentication configuration")
	}

	if !isReloading {
		return rerrors.Wrap(user_errors.ErrPgReplicationAuth)
	}

	return nil
}

func (s *pgReplicationSetup) basebackup(ctx context.Context) error {
	topology, err := s.backupTopology(ctx)
	if err != nil {
		return rerrors.Wrap(err)
	}

	if topology.target.State == nil || topology.target.State.Running {
		return rerrors.Wrap(user_errors.ErrPgReplicationTargetRunning)
	}

	source, err := s.resolve(ctx, s.payload.GetSourceInstance())
	if err != nil {
		return rerrors.Wrap(err)
	}

	name := pgBackupHelperPrefix + s.payload.GetOperationId()
	cli := s.docker.Client()
	helper, err := cli.ContainerInspect(ctx, name)

	if errdefs.IsNotFound(err) {
		helperId, createErr := s.createBackupHelper(ctx, topology, source, name)
		if createErr != nil {
			return rerrors.Wrap(createErr)
		}

		err = s.runBackupHelper(ctx, helperId, true)
		if err != nil {
			return rerrors.Wrap(err)
		}

		return nil
	}

	if err != nil {
		return rerrors.Wrap(err, "error looking up pg base backup helper")
	}

	// The helper can outlive a worker. Attach to its result instead of deleting
	// a backup in progress, or restart a helper interrupted by a forced stop.
	shouldStart := helper.State.Status == "created" || helper.State.ExitCode == pgBackupKilledExitCode

	err = s.runBackupHelper(ctx, helper.ID, shouldStart)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return nil
}

func (s *pgReplicationSetup) createBackupHelper(
	ctx context.Context, topology pgBackupTopology, source replicationEndpoint, name string,
) (string, error) {
	slot := ""
	if s.payload.GetUseSlot() {
		slot = s.slotName()
	}

	config := &container.Config{
		Image:      topology.target.Image,
		User:       pgBackupRunAsRoot,
		Entrypoint: []string{"sh", pgBackupShellOptions, pgBasebackupScript, "sh"},
		Cmd:        []string{pgReplicaDataPath, source.host, s.payload.GetOperationId(), pgReplicaMarkerFile, slot},
		Env: []string{
			"PGHOST=" + source.host,
			"PGPORT=" + strconv.Itoa(int(source.port)),
			"PGUSER=" + source.user,
			"PGPASSWORD=" + source.password,
			"PGAPPNAME=" + pgReplicaApplicationName(s.payload.GetTargetInstance()),
			"PGCONNECT_TIMEOUT=5",
		},
	}

	volumeOptions := &mount.VolumeOptions{NoCopy: true}
	volume := mount.Mount{
		Type:          mount.TypeVolume,
		Source:        topology.volume,
		Target:        pgVolumePath,
		VolumeOptions: volumeOptions,
	}
	hostConfig := &container.HostConfig{
		NetworkMode: container.NetworkMode(topology.networks[0]),
		Mounts:      []mount.Mount{volume},
	}

	created, err := s.docker.Client().ContainerCreate(ctx, config, hostConfig, nil, nil, name)
	if err != nil {
		return "", rerrors.Wrap(err, "error creating pg base backup helper")
	}

	return created.ID, nil
}

func (s *pgReplicationSetup) runBackupHelper(ctx context.Context, helperId string, shouldStart bool) error {
	cli := s.docker.Client()

	if shouldStart {
		opts := container.StartOptions{}

		err := cli.ContainerStart(ctx, helperId, opts)
		if err != nil {
			return rerrors.Wrap(err, "error starting pg base backup helper")
		}
	}

	statusCh, errorCh := cli.ContainerWait(ctx, helperId, container.WaitConditionNotRunning)

	select {
	case <-ctx.Done():
		return rerrors.Wrap(ctx.Err())
	case err := <-errorCh:
		return rerrors.Wrap(err, "error waiting for pg base backup helper")
	case status := <-statusCh:
		err := s.finishBackupHelper(ctx, helperId, status)
		if err != nil {
			return rerrors.Wrap(err)
		}

		return nil
	}
}

func (s *pgReplicationSetup) finishBackupHelper(
	ctx context.Context, helperId string, status container.WaitResponse,
) error {
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), executorRemovalTimeout)
	defer cancel()

	err := s.docker.Remove(cleanupCtx, helperId)
	if err != nil {
		return rerrors.Wrap(err, "error removing pg base backup helper")
	}

	if status.StatusCode == 0 && status.Error == nil {
		return nil
	}

	return rerrors.Wrap(user_errors.ErrPgReplicationBackupFailed)
}

func pgReplicationTargetVolume(target container.InspectResponse) string {
	if target.Config == nil || !slices.Contains(target.Config.Env, "PGDATA="+pgReplicaDataPath) {
		return ""
	}

	volume := ""

	for _, item := range target.Mounts {
		if item.Destination == pgVolumePath && item.Type == mount.TypeVolume && item.RW {
			volume = item.Name
		}

		if strings.HasPrefix(item.Destination, pgVolumePath+"/") {
			return ""
		}
	}

	return volume
}

func pgMountsVolume(source container.InspectResponse, volume string) bool {
	for _, item := range source.Mounts {
		if item.Type == mount.TypeVolume && item.Name == volume {
			return true
		}
	}

	return false
}

func pgSharedNetworks(source, target *container.NetworkSettings) []string {
	if source == nil || target == nil {
		return nil
	}

	names := make([]string, 0, len(source.Networks))

	for name, endpoint := range source.Networks {
		if endpoint != nil && target.Networks[name] != nil && isUserDefinedNetwork(name) {
			names = append(names, name)
		}
	}

	slices.Sort(names)

	return names
}

func pgReplicationSubnets(ctx context.Context, cli client.APIClient, names []string) ([]string, error) {
	var subnets []string

	for _, name := range names {
		opts := network.InspectOptions{}

		info, err := cli.NetworkInspect(ctx, name, opts)
		if err != nil {
			return nil, rerrors.Wrap(err, "error inspecting shared pg instance network")
		}

		for _, subnet := range info.IPAM.Config {
			_, ipNet, parseErr := net.ParseCIDR(subnet.Subnet)
			if parseErr != nil {
				continue
			}

			subnets = append(subnets, ipNet.String())
		}
	}

	if len(subnets) == 0 {
		return nil, rerrors.Wrap(user_errors.ErrPgReplicationNetwork)
	}

	return subnets, nil
}

func runPgReplicationExec(ctx context.Context, cli client.APIClient, id string, opts container.ExecOptions) error {
	exec, err := cli.ContainerExecCreate(ctx, id, opts)
	if err != nil {
		return rerrors.Wrap(err, "error creating pg configuration exec")
	}

	attachOpts := container.ExecAttachOptions{}

	output, err := cli.ContainerExecAttach(ctx, exec.ID, attachOpts)
	if err != nil {
		return rerrors.Wrap(err, "error attaching pg configuration exec")
	}

	defer output.Close()

	_, err = stdcopy.StdCopy(io.Discard, io.Discard, output.Reader)
	if err != nil {
		return rerrors.Wrap(err, "error waiting for pg configuration exec")
	}

	result, err := cli.ContainerExecInspect(ctx, exec.ID)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting pg configuration exec")
	}

	if result.ExitCode != 0 {
		return rerrors.Wrapf(user_errors.ErrPgReplicationAuth, "pg configuration exited with status %d", result.ExitCode)
	}

	return nil
}

func readPgReplicaMarker(ctx context.Context, cli client.APIClient, id string) (string, error) {
	content, err := readFileFromContainer(ctx, cli, id, path.Join(pgReplicaDataPath, pgReplicaMarkerFile))
	if errdefs.IsNotFound(err) {
		return "", nil
	}

	if err != nil {
		return "", rerrors.Wrap(err, "error reading replica source metadata")
	}

	source, _, hasOperation := bytes.Cut(content, []byte("\n"))
	if len(source) == 0 || !hasOperation {
		return "", rerrors.Wrap(user_errors.ErrPgReplicationMarker)
	}

	return string(source), nil
}

func hasOriginalPgData(ctx context.Context, cli client.APIClient, targetId string) (bool, error) {
	_, err := readFileFromContainer(ctx, cli, targetId, path.Join(pgReplicaDataPath, "PG_VERSION"))
	if errdefs.IsNotFound(err) {
		return false, nil
	}

	if err != nil {
		return false, rerrors.Wrap(err, "error checking original pg target data")
	}

	return true, nil
}
