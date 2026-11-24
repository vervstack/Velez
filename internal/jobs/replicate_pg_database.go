package jobs

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/pkg/stdcopy"
	_ "github.com/lib/pq"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service/secrets"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/utils/common"
)

const (
	ReplicatePgDatabaseAction = "replicate_pg_database"

	stepResolveReplicationEndpoints = "resolve_endpoints"
	stepCheckReplicationTarget      = "check_target_absent"
	stepEnsureReplicationRoles      = "ensure_roles"
	stepDumpAndRestore              = "dump_and_restore"
	stepVerifyReplication           = "verify"

	replicationExecutorImage = "postgres:18"
	executorRemovalTimeout   = 20 * time.Second
)

type replicatePgDatabaseHandler struct {
	dataStorage storage.Storage
	secrets     secrets.Store
	nodeClients node_clients.NodeClients
}

func NewReplicatePgDatabaseHandler(
	dataStorage storage.Storage, secretsStore secrets.Store, nodeClients node_clients.NodeClients,
) TaskHandler {
	return &replicatePgDatabaseHandler{dataStorage: dataStorage, secrets: secretsStore, nodeClients: nodeClients}
}

func (h *replicatePgDatabaseHandler) Action() string { return ReplicatePgDatabaseAction }

func (h *replicatePgDatabaseHandler) NewContext() TaskContext {
	return &velez_api.ReplicatePgDatabaseTaskPayload{}
}

func (h *replicatePgDatabaseHandler) BuildJobs(taskCtx TaskContext) []NamedJob {
	payload, ok := taskCtx.(*velez_api.ReplicatePgDatabaseTaskPayload)
	if !ok {
		panic("replicate_pg_database: BuildJobs called with mismatched TaskContext type")
	}

	return []NamedJob{
		{Name: stepResolveReplicationEndpoints, Job: &resolveReplicationEndpointsJob{
			services: h.dataStorage.Services(), pgInstances: h.dataStorage.PgInstances(), secrets: h.secrets, payload: payload,
		}},
		{Name: stepCheckReplicationTarget, Job: &checkReplicationTargetJob{payload: payload}},
		{Name: stepEnsureReplicationRoles, Job: &ensureReplicationRolesJob{payload: payload}},
		{Name: stepDumpAndRestore, Job: &dumpAndRestorePgDatabaseJob{docker: h.nodeClients.Docker(), payload: payload}},
		{Name: stepVerifyReplication, Job: &verifyReplicationJob{payload: payload}},
	}
}

// replicationEndpointPayload is intentionally the small accessor used by all
// jobs: mutations land on the persisted task payload before the next checkpoint.
type replicationEndpointPayload struct {
	payload *velez_api.ReplicatePgDatabaseTaskPayload
}

func (p replicationEndpointPayload) source() replicationEndpoint {
	endpoint := replicationEndpoint{
		host:     p.payload.GetSourceHost(),
		port:     p.payload.GetSourcePort(),
		user:     p.payload.GetSourceUser(),
		password: p.payload.GetSourcePassword(),
		dbName:   p.payload.GetSourceDbName(),
	}

	return endpoint
}

func (p replicationEndpointPayload) target() replicationEndpoint {
	endpoint := replicationEndpoint{
		host:     p.payload.GetTargetHost(),
		port:     p.payload.GetTargetPort(),
		user:     p.payload.GetTargetUser(),
		password: p.payload.GetTargetPassword(),
		dbName:   p.payload.GetSourceDbName(),
	}

	return endpoint
}

type replicationEndpoint struct {
	host     string
	port     uint32
	user     string
	password string
	dbName   string
}

func (e replicationEndpoint) dsn(dbName string) string {
	connection := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(e.user, e.password),
		Host:     net.JoinHostPort(e.host, strconv.Itoa(int(e.port))),
		Path:     dbName,
		RawQuery: "sslmode=disable",
	}

	return connection.String()
}

type resolveReplicationEndpointsJob struct {
	services    storage.ServicesStorage
	pgInstances storage.PgInstancesStorage
	secrets     secrets.Store
	payload     *velez_api.ReplicatePgDatabaseTaskPayload
}

func (j *resolveReplicationEndpointsJob) Do(ctx context.Context) error {
	if j.payload.GetSourceInstanceId() != "" {
		source, err := resolvePgReplicationInstance(
			ctx, j.services, j.pgInstances, j.secrets, j.payload.GetSourceInstanceId(),
		)
		if err != nil {
			return rerrors.Wrap(err, "error resolving source pg instance")
		}

		j.payload.SourceHost = source.host
		j.payload.SourcePort = source.port
		j.payload.SourceUser = source.user
		j.payload.SourcePassword = source.password
		j.payload.SourceDbName = source.dbName
	}

	if j.payload.GetTargetInstanceId() != "" {
		target, err := resolvePgReplicationInstance(
			ctx, j.services, j.pgInstances, j.secrets, j.payload.GetTargetInstanceId(),
		)
		if err != nil {
			return rerrors.Wrap(err, "error resolving target pg instance")
		}

		j.payload.TargetHost = target.host
		j.payload.TargetPort = target.port
		j.payload.TargetUser = target.user
		j.payload.TargetPassword = target.password
	}

	if j.payload.GetSourceHost() == j.payload.GetTargetHost() && j.payload.GetSourcePort() == j.payload.GetTargetPort() {
		return rerrors.New("source and target database must differ")
	}

	return nil
}

func resolvePgReplicationInstance(
	ctx context.Context,
	services storage.ServicesStorage,
	pgInstances storage.PgInstancesStorage,
	secretsStore secrets.Store,
	name string,
) (replicationEndpoint, error) {
	svc, err := services.GetByName(ctx, name)
	if err != nil {
		return replicationEndpoint{}, rerrors.Wrap(err, "error getting pg instance service")
	}

	instance, err := pgInstances.GetPgInstanceByServiceID(ctx, svc.ID)
	if err != nil {
		return replicationEndpoint{}, rerrors.Wrap(err, "error getting pg instance row")
	}

	ref, err := domain.ParseSecretRef(instance.SecretRef)
	if err != nil {
		return replicationEndpoint{}, rerrors.Wrap(err, "error parsing pg instance secret ref")
	}

	password, err := secretsStore.Get(ctx, ref)
	if err != nil {
		return replicationEndpoint{}, rerrors.Wrap(err, "error resolving pg instance password")
	}

	endpoint := replicationEndpoint{
		host:     name,
		port:     uint32(instance.Port),
		user:     instance.Username,
		password: password,
		dbName:   instance.DbName,
	}

	return endpoint, nil
}

type checkReplicationTargetJob struct {
	payload *velez_api.ReplicatePgDatabaseTaskPayload
}

func (j *checkReplicationTargetJob) Do(ctx context.Context) error {
	target := replicationEndpointPayload{payload: j.payload}.target()

	db, err := sql.Open("postgres", target.dsn("postgres"))
	if err != nil {
		return rerrors.Wrap(err, "error connecting to replication target")
	}

	defer common.CloseWithLog(db.Close, "replication target database")

	var exists bool

	err = db.QueryRowContext(
		ctx,
		"SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)",
		j.payload.GetSourceDbName(),
	).Scan(&exists)
	if err != nil {
		return rerrors.Wrap(err, "error checking replication target database")
	}

	if exists {
		return rerrors.New("target database already exists")
	}

	return nil
}

type ensureReplicationRolesJob struct {
	payload *velez_api.ReplicatePgDatabaseTaskPayload
}

func (j *ensureReplicationRolesJob) Do(ctx context.Context) error {
	source := replicationEndpointPayload{payload: j.payload}.source()
	target := replicationEndpointPayload{payload: j.payload}.target()

	sourceDb, err := sql.Open("postgres", source.dsn(source.dbName))
	if err != nil {
		return rerrors.Wrap(err, "error connecting to replication source")
	}

	defer common.CloseWithLog(sourceDb.Close, "replication source database")

	targetDb, err := sql.Open("postgres", target.dsn("postgres"))
	if err != nil {
		return rerrors.Wrap(err, "error connecting to replication target")
	}

	defer common.CloseWithLog(targetDb.Close, "replication target database")

	rows, err := sourceDb.QueryContext(ctx, `
SELECT DISTINCT r.rolname, r.rolcanlogin, r.rolinherit
FROM pg_roles r
JOIN (
    SELECT relowner AS owner_oid
    FROM pg_class
    WHERE relnamespace NOT IN (SELECT oid FROM pg_namespace WHERE nspname IN ('pg_catalog', 'information_schema'))
    UNION
    SELECT proowner AS owner_oid
    FROM pg_proc
    WHERE pronamespace NOT IN (SELECT oid FROM pg_namespace WHERE nspname IN ('pg_catalog', 'information_schema'))
    UNION
    SELECT nspowner AS owner_oid
    FROM pg_namespace
    WHERE nspname NOT IN ('pg_catalog', 'information_schema', 'pg_toast')
) owners ON owners.owner_oid = r.oid`)
	if err != nil {
		return rerrors.Wrap(err, "error listing source role owners")
	}

	defer common.CloseWithLog(rows.Close, "source role owners rows")

	for rows.Next() {
		var name string

		var canLogin, inherit bool

		err = rows.Scan(&name, &canLogin, &inherit)
		if err != nil {
			return rerrors.Wrap(err, "error reading source role owner")
		}

		var exists bool

		err = targetDb.QueryRowContext(
			ctx,
			"SELECT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = $1)",
			name,
		).Scan(&exists)
		if err != nil {
			return rerrors.Wrap(err, "error checking target role")
		}

		if exists {
			continue
		}

		login := "NOLOGIN"

		if canLogin {
			login = "LOGIN"
		}

		inheritClause := "NOINHERIT"

		if inherit {
			inheritClause = "INHERIT"
		}

		// PostgreSQL does not expose role passwords to ordinary connections; new roles intentionally have no password.
		statement := fmt.Sprintf("CREATE ROLE %s %s %s", quotePgIdentifier(name), login, inheritClause)

		_, err = targetDb.ExecContext(ctx, statement)
		if err != nil {
			return rerrors.Wrap(err, "error creating target role")
		}
	}

	err = rows.Err()
	if err != nil {
		return rerrors.Wrap(err, "error listing source role owners")
	}

	return nil
}

type dumpAndRestorePgDatabaseJob struct {
	docker  node_clients.Docker
	payload *velez_api.ReplicatePgDatabaseTaskPayload
}

func (j *dumpAndRestorePgDatabaseJob) Do(ctx context.Context) error {
	executorId, isTemporary, err := j.executorContainer(ctx)
	if err != nil {
		return rerrors.Wrap(err, "error getting replication executor")
	}

	if isTemporary {
		defer j.removeExecutor(ctx, executorId)
	}

	source := replicationEndpointPayload{payload: j.payload}.source()
	target := replicationEndpointPayload{payload: j.payload}.target()

	dumpOptions := container.ExecOptions{
		Cmd: []string{
			"pg_dump", "-Fc", "--no-owner", "-h", source.host, "-p", strconv.Itoa(int(source.port)),
			"-U", source.user, source.dbName,
		},
		Env:          []string{"PGPASSWORD=" + source.password},
		AttachStdout: true,
		AttachStderr: true,
	}

	restoreOptions := container.ExecOptions{
		Cmd: []string{
			"pg_restore", "-h", target.host, "-p", strconv.Itoa(int(target.port)), "-U", target.user,
			"-d", "postgres", "--create",
		},
		Env:          []string{"PGPASSWORD=" + target.password},
		AttachStdin:  true,
		AttachStderr: true,
	}

	cli := j.docker.Client()

	restoreExec, err := cli.ContainerExecCreate(ctx, executorId, restoreOptions)
	if err != nil {
		return rerrors.Wrap(err, "error creating pg_restore exec")
	}

	restoreAttachOptions := container.ExecAttachOptions{}

	restoreAttach, err := cli.ContainerExecAttach(ctx, restoreExec.ID, restoreAttachOptions)
	if err != nil {
		return rerrors.Wrap(err, "error attaching pg_restore exec")
	}

	defer restoreAttach.Close()

	dumpExec, err := cli.ContainerExecCreate(ctx, executorId, dumpOptions)
	if err != nil {
		return rerrors.Wrap(err, "error creating pg_dump exec")
	}

	dumpAttachOptions := container.ExecAttachOptions{}

	dumpAttach, err := cli.ContainerExecAttach(ctx, dumpExec.ID, dumpAttachOptions)
	if err != nil {
		return rerrors.Wrap(err, "error attaching pg_dump exec")
	}

	defer dumpAttach.Close()

	pipeReader, pipeWriter := io.Pipe()

	copyDone := make(chan error, 1)

	go func() {
		_, copyErr := stdcopy.StdCopy(pipeWriter, io.Discard, dumpAttach.Reader)

		copyDone <- pipeWriter.CloseWithError(copyErr)
	}()

	_, err = io.Copy(restoreAttach.Conn, pipeReader)
	if err != nil {
		return rerrors.Wrap(err, "error streaming pg_dump to pg_restore")
	}

	if closer, ok := restoreAttach.Conn.(interface{ CloseWrite() error }); ok {
		err = closer.CloseWrite()
		if err != nil {
			return rerrors.Wrap(err, "error closing pg_restore input")
		}
	}

	// The attach stream closes only when the exec finishes - drain it so the
	// exit-code inspect below can't observe pg_restore still running.
	_, err = io.Copy(io.Discard, restoreAttach.Reader)
	if err != nil {
		return rerrors.Wrap(err, "error waiting for pg_restore to finish")
	}

	err = <-copyDone
	if err != nil {
		return rerrors.Wrap(err, "error reading pg_dump stream")
	}

	err = j.checkExec(ctx, dumpExec.ID, "pg_dump", cli, restoreExec.ID, "pg_restore")
	if err != nil {
		return rerrors.Wrap(err, "error checking replication commands")
	}

	return nil
}

func (j *dumpAndRestorePgDatabaseJob) executorContainer(ctx context.Context) (string, bool, error) {
	cli := j.docker.Client()

	if j.payload.GetSourceInstanceId() != "" {
		args := filters.NewArgs(filters.Arg("label", labels.VervServiceLabel+"="+j.payload.GetSourceInstanceId()))
		listOptions := container.ListOptions{Filters: args}

		containers, err := cli.ContainerList(ctx, listOptions)
		if err != nil {
			return "", false, rerrors.Wrap(err, "error listing source instance container")
		}

		if len(containers) != 0 {
			return containers[0].ID, false, nil
		}
	}

	args := filters.NewArgs(filters.Arg("label", labels.VervServiceLabel))
	listOptions := container.ListOptions{Filters: args}

	containers, err := cli.ContainerList(ctx, listOptions)
	if err != nil {
		return "", false, rerrors.Wrap(err, "error listing pgaas executor containers")
	}

	for _, candidate := range containers {
		if strings.HasPrefix(candidate.Labels[labels.VervServiceLabel], labels.PgaasNamePrefix) {
			return candidate.ID, false, nil
		}
	}

	_, err = j.docker.PullImage(ctx, replicationExecutorImage)
	if err != nil {
		return "", false, rerrors.Wrap(err, "error pulling replication executor image")
	}

	config := &container.Config{Image: replicationExecutorImage, Cmd: []string{"sleep", "infinity"}}

	created, err := cli.ContainerCreate(ctx, config, nil, nil, nil, "")
	if err != nil {
		return "", false, rerrors.Wrap(err, "error creating replication executor")
	}

	startOptions := container.StartOptions{}

	err = cli.ContainerStart(ctx, created.ID, startOptions)
	if err != nil {
		return "", false, rerrors.Wrap(err, "error starting replication executor")
	}

	return created.ID, true, nil
}

func (j *dumpAndRestorePgDatabaseJob) removeExecutor(ctx context.Context, id string) {
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), executorRemovalTimeout)

	defer cancel()

	_ = j.docker.Remove(cleanupCtx, id)
}

func (j *dumpAndRestorePgDatabaseJob) checkExec(ctx context.Context, firstId, firstName string, cli interface {
	ContainerExecInspect(ctx context.Context, execId string) (container.ExecInspect, error)
}, secondId, secondName string) error {
	first, err := cli.ContainerExecInspect(ctx, firstId)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting "+firstName)
	}

	if first.ExitCode != 0 {
		return rerrors.New(firstName + " failed")
	}

	second, err := cli.ContainerExecInspect(ctx, secondId)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting "+secondName)
	}

	if second.ExitCode != 0 {
		return rerrors.New(secondName + " failed")
	}

	return nil
}

type verifyReplicationJob struct {
	payload *velez_api.ReplicatePgDatabaseTaskPayload
}

func (j *verifyReplicationJob) Do(ctx context.Context) error {
	source := replicationEndpointPayload{payload: j.payload}.source()
	target := replicationEndpointPayload{payload: j.payload}.target()

	sourceDb, err := sql.Open("postgres", source.dsn(source.dbName))
	if err != nil {
		return rerrors.Wrap(err, "error connecting to replication source")
	}

	defer common.CloseWithLog(sourceDb.Close, "replication source database")

	targetDb, err := sql.Open("postgres", target.dsn(source.dbName))
	if err != nil {
		return rerrors.Wrap(err, "error connecting to replication target")
	}

	defer common.CloseWithLog(targetDb.Close, "replication target database")

	rows, err := sourceDb.QueryContext(
		ctx,
		`SELECT schemaname, relname FROM pg_stat_user_tables ORDER BY schemaname, relname`,
	)
	if err != nil {
		return rerrors.Wrap(err, "error listing source tables")
	}

	defer common.CloseWithLog(rows.Close, "source tables rows")

	for rows.Next() {
		var schema, table string

		err = rows.Scan(&schema, &table)
		if err != nil {
			return rerrors.Wrap(err, "error reading source table")
		}

		statement := "SELECT count(*) FROM " + quotePgIdentifier(schema) + "." + quotePgIdentifier(table)

		var sourceCount, targetCount int64

		err = sourceDb.QueryRowContext(ctx, statement).Scan(&sourceCount)
		if err != nil {
			return rerrors.Wrap(err, "error counting source rows")
		}

		err = targetDb.QueryRowContext(ctx, statement).Scan(&targetCount)
		if err != nil {
			return rerrors.Wrap(err, "error counting target rows")
		}

		if sourceCount != targetCount {
			return rerrors.New("replication verification failed: table row count differs")
		}
	}

	err = rows.Err()
	if err != nil {
		return rerrors.Wrap(err, "error listing source tables")
	}

	err = verifySequenceValues(ctx, sourceDb, targetDb)
	if err != nil {
		return rerrors.Wrap(err, "error verifying sequence values")
	}

	return nil
}

func verifySequenceValues(ctx context.Context, sourceDb, targetDb *sql.DB) error {
	rows, err := sourceDb.QueryContext(
		ctx,
		`SELECT schemaname, sequencename FROM pg_sequences ORDER BY schemaname, sequencename`,
	)
	if err != nil {
		return rerrors.Wrap(err, "error listing source sequences")
	}

	defer common.CloseWithLog(rows.Close, "source sequences rows")

	for rows.Next() {
		var schema, sequence string

		err = rows.Scan(&schema, &sequence)
		if err != nil {
			return rerrors.Wrap(err, "error reading source sequence")
		}

		statement := "SELECT last_value FROM " + quotePgIdentifier(schema) + "." + quotePgIdentifier(sequence)

		var sourceValue, targetValue int64

		err = sourceDb.QueryRowContext(ctx, statement).Scan(&sourceValue)
		if err != nil {
			return rerrors.Wrap(err, "error reading source sequence value")
		}

		err = targetDb.QueryRowContext(ctx, statement).Scan(&targetValue)
		if err != nil {
			return rerrors.Wrap(err, "error reading target sequence value")
		}

		if sourceValue != targetValue {
			return rerrors.New("replication verification failed: sequence value differs")
		}
	}

	err = rows.Err()
	if err != nil {
		return rerrors.Wrap(err, "error listing source sequences")
	}

	return nil
}

func quotePgIdentifier(value string) string {
	quoted := `"` + strings.ReplaceAll(value, `"`, `""`) + `"`

	return quoted
}
