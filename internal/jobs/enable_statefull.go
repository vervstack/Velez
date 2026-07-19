package jobs

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/errdefs"
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
	"go.vervstack.ru/Velez/internal/patterns/db_patterns/pg_pattern"
	"go.vervstack.ru/Velez/internal/pipelines/steps/cluster_steps"
	"go.vervstack.ru/Velez/internal/storage"
)

const EnableStatefullAction = "enable_statefull_mode"

const (
	pgSchema                = "velez"
	pgMasterNodeDefaultName = "icy_raccoon"
)

// Accessor interfaces the enable_statefull jobs need from their TaskContext.
// *velez_api.EnableStatefullTaskPayload satisfies all of them.
// containerIdAccessor is declared in create_smerd.go and reused here as-is.

type statefullRequestAccessor interface {
	GetRequest() *velez_api.EnableStatefullCluster
}

type rootPwdAccessor interface {
	GetRootPwd() string
	SetRootPwd(string)
}

type userPwdAccessor interface {
	GetUserPwd() string
	SetUserPwd(string)
}

type rootDsnAccessor interface {
	GetRootDsn() string
	SetRootDsn(string)
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
}

func NewEnableStatefullHandler(
	nodeClients node_clients.NodeClients,
	clusterStateManager cluster_clients.ClusterStateManagerContainer,
	storageContainer *storage.Container,
) TaskHandler {
	return &enableStatefullHandler{
		nodeClients:         nodeClients,
		clusterStateManager: clusterStateManager,
		storageContainer:    storageContainer,
	}
}

func (h *enableStatefullHandler) Action() string {
	return EnableStatefullAction
}

func (h *enableStatefullHandler) NewContext() TaskContext {
	return &velez_api.EnableStatefullTaskPayload{}
}

// BuildJobs mirrors do_enable_statefull.go's 7 pipeline steps, plus one extra
// job (generate_credentials) promoted out of the pipeline's inline "Pipeline
// Context" setup - see docs/jobs_migrations/questions.md for why that setup
// can't stay inline BuildJobs code: it derives passwords by checking whether
// they're already persisted in local/cluster state, which for jobs must be
// checked against the task's own persisted context instead, exactly once,
// guarded by a checkpoint (same idempotency precedent as
// ConnectServiceToVpnTaskPayload.client_key).
func (h *enableStatefullHandler) BuildJobs(taskCtx TaskContext) []NamedJob {
	payload := taskCtx.(*velez_api.EnableStatefullTaskPayload)

	return []NamedJob{
		{
			Name: "generate_credentials",
			Job: &generateCredentialsJob{
				localStateManager: h.nodeClients.LocalStateManager(),
				ctx:               payload,
			},
		},
		{
			Name: "create_container",
			Job: &createPgContainerJob{
				nodeClients: h.nodeClients,
				req:         payload,
				pwd:         payload,
				ctx:         payload,
			},
		},
		{
			Name: "start_container",
			Job: &startPgContainerJob{
				dockerAPI: h.nodeClients.Docker().Client(),
				ctx:       payload,
			},
		},
		{
			Name: "get_root_dsn",
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
			rootPwd = string(toolbox.RandomBase64(12))
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
			userPwd = string(toolbox.RandomBase64(12))
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
	ctx containerIdAccessor
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
	containerId := j.ctx.GetContainerId()
	if containerId == "" {
		return nil
	}

	err := j.nodeClients.Docker().Remove(ctx, containerId)
	if err != nil && !errdefs.IsNotFound(err) {
		return rerrors.Wrapf(err, "error removing postgres container '%s'", containerId)
	}

	return nil
}

// startPgContainerJob mirrors smerd_steps.Start, reusing the containerIdAccessor
// and startAPI-shaped dependency already established for create_smerd.go/
// connect_service_to_vpn.go.
type startPgContainerJob struct {
	dockerAPI startAPI

	ctx containerIdAccessor
}

func (j *startPgContainerJob) Do(ctx context.Context) error {
	containerId := j.ctx.GetContainerId()
	if containerId == "" {
		return rerrors.New("no container id provided")
	}

	err := j.dockerAPI.ContainerStart(ctx, containerId, container.StartOptions{})
	if err != nil {
		return rerrors.Wrap(err, "error starting postgres container")
	}

	return nil
}

func (j *startPgContainerJob) Rollback(ctx context.Context) error {
	containerId := j.ctx.GetContainerId()
	if containerId == "" {
		return nil
	}

	err := j.dockerAPI.ContainerStop(ctx, containerId, container.StopOptions{})
	if err != nil {
		return rerrors.Wrapf(err, "error stopping postgres container '%s'", containerId)
	}

	return nil
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
		containerIdAccessor
		rootDsnAccessor
	}
}

func (j *getRootDsnJob) Do(ctx context.Context) error {
	containerId := j.ctx.GetContainerId()

	pgCfg := &resources.Postgres{
		Host: state.PgName,
		Port: 5432,

		User:    "postgres",
		Pwd:     "",
		DbName:  "",
		SslMode: "disable",
	}

	cont, err := j.dockerAPI.ContainerInspect(ctx, containerId)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting postgres container")
	}

	for _, v := range cont.Config.Env {
		envVarParts := strings.SplitN(v, "=", 2)
		if len(envVarParts) < 2 {
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
			return rerrors.Wrap(err)
		}
	}

	j.ctx.SetRootDsn(pgCfg.ConnectionString() + "&application_name=RootSetup")

	return nil
}

func getExposedPgPort(cont container.InspectResponse) (uint64, error) {
	if cont.NetworkSettings == nil {
		return 0, rerrors.New("no network settings found in container")
	}

	ports := cont.NetworkSettings.Ports[pg_pattern.TcpPort]
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
// migration. It reuses sqldb.RollMigration verbatim, so it inherits that
// function's existing (already untested at unit level, pre-migration) real
// Postgres connection requirement - see docs/jobs_migrations/questions.md.
type createSchemaAndMigrateJob struct {
	dsn rootDsnAccessor
}

func (j *createSchemaAndMigrateJob) Do(_ context.Context) error {
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

	_, err = conn.Exec(`CREATE SCHEMA IF NOT EXISTS velez`)
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

	step := cluster_steps.CreatePgUserForNode(&rootDsn, pgSchema, pgMasterNodeDefaultName, j.pwd.GetUserPwd())

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
}

func (j *initNodeStorageJob) Do(ctx context.Context) error {
	err := j.clusterStateManager.Nodes().InitNode(ctx)
	if err != nil {
		return rerrors.Wrap(err, "error initializing node storage")
	}

	return nil
}
