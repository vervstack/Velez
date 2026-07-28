package jobs

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"
	"go.vervstack.ru/matreshka/pkg/matreshka/resources"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients/state"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/sqldb"
	"go.vervstack.ru/Velez/internal/cluster/env"
	"go.vervstack.ru/Velez/internal/config"
	"go.vervstack.ru/Velez/internal/patterns/db_patterns/pg_pattern"
	"go.vervstack.ru/Velez/internal/pipelines/steps/cluster_steps"
	"go.vervstack.ru/Velez/internal/storage"
)

const (
	EnableStatefullAction = "enable_statefull_mode"

	pgMasterNodeDefaultName = "icy_raccoon"

	stepGenerateCredentials  = "generate_credentials"
	stepCreatePgContainer    = "create_container"
	stepWaitForPostgresReady = "wait_for_postgres_ready"
	stepGetRootDsn           = "get_root_dsn"

	pgDefaultUser = "postgres"

	// generatedPwdLength is the byte length passed to toolbox.RandomBase64 when
	// generating a fresh root/user postgres password.
	generatedPwdLength = 12

	pgDefaultPort   = 5432
	envKVPartsCount = 2

	// pgReadyPollInterval/pgReadyTimeout bound waitForPgReadyJob's poll loop. A
	// freshly started postgres:18 container runs initdb and restarts itself
	// before it truly accepts connections - observed ~6s on a quiet local Docker
	// host - closing any connection attempted during that window with a bare
	// EOF, which is what create_schema_and_migrate saw when this wait step
	// didn't exist yet. 30s gives ample margin on a loaded host without letting
	// a genuinely wedged container hang the task indefinitely.
	pgReadyPollInterval = 500 * time.Millisecond
	pgReadyTimeout      = 30 * time.Second
)

// Accessor interfaces the enable_statefull jobs need from their TaskContext.
// *velez_api.EnableStatefullTaskPayload satisfies all of them.
// containerIDAccessor is declared in create_smerd.go and reused here as-is.

type statefullRequestAccessor interface {
	GetRequest() *velez_api.EnableStatefullCluster
}

type rootPwdAccessor interface {
	GetRootPwd() string
	SetRootPwd(rootPwd string)
}

type userPwdAccessor interface {
	GetUserPwd() string
	SetUserPwd(userPwd string)
}

type rootDsnAccessor interface {
	GetRootDsn() string
	SetRootDsn(rootDsn string)
}

// containerInspectAPI is the narrow docker slice getRootDsnJob depends on
// (ContainerInspect only), following the startAPI/copyAPI precedent from
// copy_to_volume.go/connect_service_to_vpn.go so the job is unit-testable
// without the full (100+ method) client.APIClient interface.
type containerInspectAPI interface {
	ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error)
}

type enableStatefullHandler struct {
	nodeClients         node_clients.NodeClients
	clusterStateManager cluster_clients.ClusterStateManagerContainer
	storageContainer    *storage.Container
	cfg                 config.Config
}

func NewEnableStatefullHandler(
	nodeClients node_clients.NodeClients,
	clusterStateManager cluster_clients.ClusterStateManagerContainer,
	storageContainer *storage.Container,
	cfg config.Config,
) TaskHandler {
	return &enableStatefullHandler{
		nodeClients:         nodeClients,
		clusterStateManager: clusterStateManager,
		storageContainer:    storageContainer,
		cfg:                 cfg,
	}
}

func (h *enableStatefullHandler) Action() string {
	return EnableStatefullAction
}

func (h *enableStatefullHandler) NewContext() TaskContext {
	return &velez_api.EnableStatefullTaskPayload{}
}

// BuildJobs mirrors do_enable_statefull.go's 7 pipeline steps, plus two extra
// jobs. generate_credentials is promoted out of the pipeline's inline
// "Pipeline Context" setup - see docs/jobs_migrations/questions.md for why
// that setup can't stay inline BuildJobs code: it derives passwords by
// checking whether they're already persisted in local/cluster state, which
// for jobs must be checked against the task's own persisted context instead,
// exactly once, guarded by a checkpoint (same idempotency precedent as
// ConnectServiceToVpnTaskPayload.client_key). wait_for_postgres_ready has no
// pipeline-step equivalent at all: it was added after tests/e2e caught this
// job chain racing a freshly started postgres container (see
// waitForPgReadyJob's doc comment).
func (h *enableStatefullHandler) BuildJobs(taskCtx TaskContext) []NamedJob {
	payload, ok := taskCtx.(*velez_api.EnableStatefullTaskPayload)
	if !ok {
		panic("enable_statefull: BuildJobs called with mismatched TaskContext type")
	}

	return []NamedJob{
		{
			Name: stepGenerateCredentials,
			Job: &generateCredentialsJob{
				localStateManager: h.nodeClients.LocalStateManager(),
				ctx:               payload,
			},
		},
		{
			Name: stepCreatePgContainer,
			Job: &createPgContainerJob{
				nodeClients: h.nodeClients,
				req:         payload,
				pwd:         payload,
				ctx:         payload,
			},
		},
		{
			Name: stepStartSidecar,
			Job: &startPgContainerJob{
				dockerAPI: h.nodeClients.Docker().Client(),
				ctx:       payload,
			},
		},
		{
			Name: stepWaitForPostgresReady,
			Job: &waitForPgReadyJob{
				dockerAPI: h.nodeClients.Docker().Client(),
				ctx:       payload,
			},
		},
		{
			Name: stepGetRootDsn,
			Job: &getRootDsnJob{
				dockerAPI: h.nodeClients.Docker().Client(),
				req:       payload,
				ctx:       payload,
			},
		},
		{
			Name: "create_schema_and_migrate",
			Job: &createSchemaAndMigrateJob{
				dsn: payload,
			},
		},
		{
			Name: "create_pg_user",
			Job: &createPgUserJob{
				dsn: payload,
				pwd: payload,
			},
		},
		{
			Name: "update_cluster_state",
			Job: &updateClusterStateJob{
				localStateManager:   h.nodeClients.LocalStateManager(),
				clusterStateManager: h.clusterStateManager,
				storageContainer:    h.storageContainer,
				newPgStateManager:   state.NewPgStateManager,
				dsn:                 payload,
				pwd:                 payload,
			},
		},
		{
			Name: "init_node_storage",
			Job: &initNodeStorageJob{
				clusterStateManager: h.clusterStateManager,
				region:              h.cfg.Environment.NodeRegion,
			},
		},
	}
}

// generateCredentialsJob derives the postgres root/node user passwords the
// same way do_enable_statefull.go's setup region does (reuse the previous
// dsn's password if one's parseable, else mint a random one), but persists
// them into the task's context the first time so resuming a partially-done
// task reuses the same passwords instead of regenerating ones that would
// mismatch an already-created container/user.
type generateCredentialsJob struct {
	localStateManager node_clients.StateManager

	ctx interface {
		rootPwdAccessor
		userPwdAccessor
	}
}

func (j *generateCredentialsJob) Do(_ context.Context) error {
	if j.ctx.GetRootPwd() == "" {
		localState := j.localStateManager.Get()

		var rootPwd string

		rootPg := resources.Postgres{}
		if rootPg.ParseFromDsn(localState.ClusterState.PgRootDsn) == nil {
			rootPwd = rootPg.Pwd
		} else {
			rootPwd = string(toolbox.RandomBase64(generatedPwdLength))
		}

		j.ctx.SetRootPwd(rootPwd)
	}

	if j.ctx.GetUserPwd() == "" {
		localState := j.localStateManager.Get()

		var userPwd string

		userPg := resources.Postgres{}
		if userPg.ParseFromDsn(localState.ClusterState.PgNodeDsn) == nil {
			userPwd = userPg.Pwd
		} else {
			userPwd = string(toolbox.RandomBase64(generatedPwdLength))
		}

		j.ctx.SetUserPwd(userPwd)
	}

	return nil
}

// createPgContainerJob mirrors do_enable_statefull.go's container_steps.Create
// call, restricted to what the postgres container needs. Like
// createSidecarContainerJob/create_smerd.go's createContainerJob (see
// questions.md #13's precedent), it doesn't tolerate an existing
// same-named container on create or clean up volume mounts on rollback -
// simpler create+Remove, consistent with how this migration already treats
// container create/rollback elsewhere.
type createPgContainerJob struct {
	nodeClients node_clients.NodeClients

	req statefullRequestAccessor
	pwd rootPwdAccessor
	ctx containerIDAccessor
}

func (j *createPgContainerJob) Do(ctx context.Context) error {
	req := j.req.GetRequest()

	var ops []pg_pattern.Opt

	ops = append(ops, pg_pattern.WithInstanceName(state.PgName))

	if req.GetIsExposePort() {
		if req.GetExposeToPort() != 0 {
			ops = append(ops, pg_pattern.WithPort(req.GetExposeToPort()))
		} else {
			ops = append(ops, pg_pattern.WithExposedPort())
		}
	}

	ops = append(ops, pg_pattern.WithPassword(j.pwd.GetRootPwd()))

	launchContainer := pg_pattern.Postgres(ops...)

	dockerClient := j.nodeClients.Docker()

	created, err := dockerClient.ContainerCreate(ctx,
		launchContainer.Pattern.Config,
		launchContainer.Pattern.HostConfig,
		launchContainer.Pattern.NetworkingConfig,
		&v1.Platform{}, state.PgName)
	if err != nil {
		return rerrors.Wrap(err, "error creating postgres container")
	}

	j.ctx.SetContainerId(created.ID)

	return nil
}

func (j *createPgContainerJob) Rollback(ctx context.Context) error {
	containerID := j.ctx.GetContainerId()
	if containerID == "" {
		return nil
	}

	err := j.nodeClients.Docker().Remove(ctx, containerID)
	if err != nil && !errdefs.IsNotFound(err) {
		return rerrors.Wrapf(err, "error removing postgres container '%s'", containerID)
	}

	return nil
}

// startPgContainerJob mirrors smerd_steps.Start, reusing the containerIDAccessor
// and startAPI-shaped dependency already established for create_smerd.go/
// connect_service_to_vpn.go.
type startPgContainerJob struct {
	dockerAPI startAPI

	ctx containerIDAccessor
}

func (j *startPgContainerJob) Do(ctx context.Context) error {
	containerID := j.ctx.GetContainerId()
	if containerID == "" {
		return rerrors.New("no container id provided")
	}

	err := j.dockerAPI.ContainerStart(ctx, containerID, container.StartOptions{})
	if err != nil {
		return rerrors.Wrap(err, "error starting postgres container")
	}

	return nil
}

func (j *startPgContainerJob) Rollback(ctx context.Context) error {
	containerID := j.ctx.GetContainerId()
	if containerID == "" {
		return nil
	}

	err := j.dockerAPI.ContainerStop(ctx, containerID, container.StopOptions{})
	if err != nil {
		return rerrors.Wrapf(err, "error stopping postgres container '%s'", containerID)
	}

	return nil
}

// waitForPgReadyJob polls the postgres container's Docker healthcheck
// (pg_isready, defined by pg_pattern.Postgres) until it reports "healthy" or
// pgReadyTimeout elapses. Without this step, create_schema_and_migrate
// dials the container immediately after ContainerStart returns, which races
// postgres's initdb-then-restart sequence and fails with a bare EOF on
// essentially every cold start - caught by tests/e2e's EnableStatefullSuite.
// No Rollback: this job only observes the container it doesn't own, mirroring
// checkSelfUpgradeJob's read-only, non-rollbackable shape.
type waitForPgReadyJob struct {
	dockerAPI containerInspectAPI

	ctx containerIDAccessor
}

func (j *waitForPgReadyJob) Do(ctx context.Context) error {
	containerID := j.ctx.GetContainerId()
	if containerID == "" {
		return rerrors.New("no container id provided")
	}

	deadline := time.Now().Add(pgReadyTimeout)

	for {
		cont, err := j.dockerAPI.ContainerInspect(ctx, containerID)
		if err != nil {
			return rerrors.Wrap(err, "error inspecting postgres container while waiting for readiness")
		}

		isHealthy := cont.ContainerJSONBase != nil && cont.State != nil && cont.State.Health != nil &&
			cont.State.Health.Status == container.Healthy
		if isHealthy {
			return nil
		}

		if time.Now().After(deadline) {
			return rerrors.New("timed out waiting for postgres container to become healthy")
		}

		select {
		case <-ctx.Done():
			return rerrors.Wrap(ctx.Err(), "wait for postgres ready context done")
		case <-time.After(pgReadyPollInterval):
		}
	}
}

// getRootDsnJob duplicates cluster_steps.GetRgRootDsn's inspect/env-parse
// logic against the narrow containerInspectAPI instead of depending on
// GetRgRootDsn's steps.Step value directly (which requires the full
// node_clients.Docker interface for its ContainerInspect call) - same
// narrow-interface-over-duplication tradeoff as copy_to_volume.go's
// writeFileToContainer helper (questions.md #8): a fake satisfying only
// containerInspectAPI can never be passed where node_clients.Docker is
// expected, so reusing the step as-is isn't possible without giving up unit
// testability.
type getRootDsnJob struct {
	dockerAPI containerInspectAPI

	req statefullRequestAccessor
	ctx interface {
		containerIDAccessor
		rootDsnAccessor
	}
}

func (j *getRootDsnJob) Do(ctx context.Context) error {
	containerID := j.ctx.GetContainerId()

	pgCfg := &resources.Postgres{
		Host: state.PgName,
		Port: pgDefaultPort,

		User:    pgDefaultUser,
		Pwd:     "",
		DbName:  "",
		SslMode: "disable",
	}

	cont, err := j.dockerAPI.ContainerInspect(ctx, containerID)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting postgres container")
	}

	for _, v := range cont.Config.Env {
		envVarParts := strings.SplitN(v, "=", envKVPartsCount)
		if len(envVarParts) < envKVPartsCount {
			continue
		}

		switch envVarParts[0] {
		case pg_pattern.DbEnvVariable:
			pgCfg.DbName = envVarParts[1]
		case pg_pattern.UserEnvVariable:
			pgCfg.User = envVarParts[1]
		case pg_pattern.PasswordEnvVariable:
			pgCfg.Pwd = envVarParts[1]
		default:
			continue
		}
	}

	isPortExposed := j.req.GetRequest().GetIsExposePort()

	if !env.IsInContainer() {
		if !isPortExposed {
			return rerrors.New("when running velez as a binary - the created container ports must be exposed")
		}

		pgCfg.Host = "localhost"

		pgCfg.Port, err = getExposedPgPort(cont)
		if err != nil {
			return rerrors.Wrap(err, "error getting exposed pg port")
		}
	}

	j.ctx.SetRootDsn(pgCfg.ConnectionString() + "&application_name=RootSetup")

	return nil
}

func getExposedPgPort(cont container.InspectResponse) (uint64, error) {
	if cont.NetworkSettings == nil {
		return 0, rerrors.New("no network settings found in container")
	}

	ports := cont.NetworkSettings.Ports[pg_pattern.TCPPort]
	if len(ports) == 0 {
		return 0, rerrors.New("no exposure for 5432 found")
	}

	hostPort := ports[0].HostPort

	port, err := strconv.ParseUint(hostPort, 10, 64)
	if err != nil {
		return 0, rerrors.Wrap(err, "error parsing exposed port for container to uint64")
	}

	return port, nil
}

// createSchemaAndMigrateJob mirrors do_enable_statefull.go's inline
// steps.SingleFunc that creates the velez schema and rolls the goose
// migration. The pre-emptive `CREATE SCHEMA IF NOT EXISTS velez` is required,
// not redundant: goose.Up (sqldb.RollMigration) bootstraps its own
// schema-qualified tracking table ("velez.__migrations", see
// sqldb.RollMigration's SetTableName call) before it runs any migration
// file, so the "velez" schema must already exist or that bootstrap itself
// fails with "schema \"velez\" does not exist" - confirmed against a real,
// truly fresh Postgres via tests/e2e's EnableStatefullSuite, which is the
// first thing to have ever exercised this path end to end.
// migrations/20251221072452_initial.sql's own `CREATE SCHEMA velez;` was
// changed to `CREATE SCHEMA IF NOT EXISTS velez;` to stop conflicting with
// this pre-create (see that file's comment).
type createSchemaAndMigrateJob struct {
	dsn rootDsnAccessor
}

func (j *createSchemaAndMigrateJob) Do(ctx context.Context) error {
	rootDsn := j.dsn.GetRootDsn()

	conn, err := sql.Open(sqldb.Dialect, rootDsn)
	if err != nil {
		return rerrors.Wrap(err, "error checking connection to postgres")
	}

	defer func() {
		closeErr := conn.Close()
		if closeErr != nil {
			log.Error().Err(closeErr).Msg("error closing postgres root connection when rolling prep migration")
		}
	}()

	_, err = conn.ExecContext(ctx, `CREATE SCHEMA IF NOT EXISTS velez`)
	if err != nil {
		return rerrors.Wrap(err, "error creating postgres schema")
	}

	err = sqldb.RollMigration(rootDsn)
	if err != nil {
		return rerrors.Wrap(err, "error rolling migration")
	}

	return nil
}

// createPgUserJob wraps cluster_steps.CreatePgUserForNode the same way
// getRootDsnJob wraps GetRgRootDsn: the dsn/password it needs are only
// available from the task's persisted context at Do()-time, not at
// BuildJobs-construction time.
type createPgUserJob struct {
	dsn rootDsnAccessor
	pwd userPwdAccessor
}

func (j *createPgUserJob) Do(ctx context.Context) error {
	rootDsn := j.dsn.GetRootDsn()

	step := cluster_steps.CreatePgUserForNode(&rootDsn, pgMasterNodeDefaultName, j.pwd.GetUserPwd())

	err := step.Do(ctx)
	if err != nil {
		return rerrors.Wrap(err, "error creating postgres user for node")
	}

	return nil
}

// updateClusterStateJob mirrors do_enable_statefull.go's final SingleFunc:
// it persists the resolved dsns into local state and swaps the live
// ClusterStateManager/StorageContainer singletons over to the freshly
// bootstrapped Postgres-backed storage - carried over as-is from the
// pipeline (changing *how* that swap happens is out of scope for a
// behavior-preserving migration, flagged in docs/jobs_migrations/questions.md).
// newPgStateManager is an injected seam (defaults to state.NewPgStateManager
// in BuildJobs) purely so this job's state-swap logic is unit-testable
// without opening a real Postgres connection.
type updateClusterStateJob struct {
	localStateManager   node_clients.StateManager
	clusterStateManager cluster_clients.ClusterStateManagerContainer
	storageContainer    *storage.Container

	newPgStateManager func(ctx context.Context, dsn string) (cluster_clients.ClusterStateManager, error)

	dsn rootDsnAccessor
	pwd userPwdAccessor
}

func (j *updateClusterStateJob) Do(ctx context.Context) error {
	localState := j.localStateManager.GetForUpdate()

	rootDsn := j.dsn.GetRootDsn()

	localState.ClusterState.PgRootDsn = rootDsn

	var nodeConnection resources.Postgres

	err := nodeConnection.ParseFromDsn(rootDsn)
	if err != nil {
		return rerrors.Wrap(err, "error parsing root user database connection")
	}

	nodeConnection.User = pgMasterNodeDefaultName
	nodeConnection.Pwd = j.pwd.GetUserPwd()

	localState.ClusterState.PgNodeDsn = nodeConnection.ConnectionString() + "&application_name=" + pgMasterNodeDefaultName

	pgClusterState, err := j.newPgStateManager(ctx, localState.ClusterState.PgNodeDsn)
	if err != nil {
		return rerrors.Wrap(err, "error initializing pgClusterState")
	}

	j.clusterStateManager.Set(pgClusterState)
	j.storageContainer.Set(pgClusterState)

	j.localStateManager.SetAndRelease(localState)

	return nil
}

// initNodeStorageJob mirrors do_enable_statefull.go's final step, run after
// update_cluster_state has swapped the live ClusterStateManager over to the
// new Postgres-backed one.
type initNodeStorageJob struct {
	clusterStateManager cluster_clients.ClusterStateManagerContainer
	region              string
}

func (j *initNodeStorageJob) Do(ctx context.Context) error {
	err := j.clusterStateManager.Nodes().InitNode(ctx, j.region)
	if err != nil {
		return rerrors.Wrap(err, "error initializing node storage")
	}

	return nil
}
