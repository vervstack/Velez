package jobs

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/errdefs"
	"github.com/docker/go-connections/nat"
	"github.com/sqlc-dev/pqtype"
	"go.vervstack.ru/matreshka/pkg/matreshka/resources"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients/state"
	"go.vervstack.ru/Velez/internal/clients/node_clients/local_state"
	"go.vervstack.ru/Velez/internal/cluster/env"
	"go.vervstack.ru/Velez/internal/config"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/jobs_queries"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	testPgContainerID = "pg123"
)

var (
	errDbUnreachable       = errors.New("db unreachable")
	errConnectionRefused   = errors.New("connection refused")
	errNoSpaceLeftOnDevice = errors.New("no space left on device")
)

func TestEnableStatefullHandler_Action(t *testing.T) {
	h := NewEnableStatefullHandler(nil, nil, nil, config.Config{})

	if h.Action() != EnableStatefullAction {
		t.Errorf("expected action %q, got %q", EnableStatefullAction, h.Action())
	}
}

func TestEnableStatefullHandler_NewContext(t *testing.T) {
	h := NewEnableStatefullHandler(nil, nil, nil, config.Config{})

	if _, ok := h.NewContext().(*velez_api.EnableStatefullTaskPayload); !ok {
		t.Fatal("expected NewContext to return *velez_api.EnableStatefullTaskPayload")
	}
}

func TestEnableStatefullHandler_BuildJobs_NamesAndOrder(t *testing.T) {
	payload := &velez_api.EnableStatefullTaskPayload{Request: &velez_api.EnableStatefullCluster{}}

	docker := newFakeDocker()
	nodeClients := newFakeNodeClients(docker)

	nodeClients.localState = newFakeStateManager(local_state.State{})

	clusterStorage := &fakeClusterStorage{nodes: &fakeNodesStorage{}}
	clusterStateManager := state.NewContainer(clusterStorage)
	storageContainer := storage.NewStorageContainer(clusterStorage)

	h := NewEnableStatefullHandler(nodeClients, clusterStateManager, storageContainer, config.Config{})

	namedJobs := h.BuildJobs(payload)

	wantNames := []string{
		stepGenerateCredentials, stepCreatePgContainer, stepStartSidecar, "wait_for_postgres_ready",
		stepGetRootDsn, "create_schema_and_migrate", "create_pg_user", "update_cluster_state",
		"init_node_storage",
	}
	if len(namedJobs) != len(wantNames) {
		t.Fatalf("expected %d jobs, got %d", len(wantNames), len(namedJobs))
	}

	for i, name := range wantNames {
		if namedJobs[i].Name != name {
			t.Errorf("expected job %d named %q, got %q", i, name, namedJobs[i].Name)
		}
	}
}

// generateCredentialsJob

func TestGenerateCredentialsJob_NoExistingState_GeneratesRandomPasswords(t *testing.T) {
	localState := newFakeStateManager(local_state.State{})
	payload := &velez_api.EnableStatefullTaskPayload{}

	j := &generateCredentialsJob{localStateManager: localState, ctx: payload}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if payload.GetRootPwd() == "" {
		t.Error("expected a root password to be generated")
	}

	if payload.GetUserPwd() == "" {
		t.Error("expected a user password to be generated")
	}
}

func TestGenerateCredentialsJob_ExistingDsn_ReusesPassword(t *testing.T) {
	rootPg := resources.Postgres{
		Host: "h", Port: 5432, User: pgDefaultUser, Pwd: "existing-root-pwd", DbName: pgDefaultUser,
	}
	nodePg := resources.Postgres{
		Host: "h", Port: 5432, User: "icy_raccoon", Pwd: "existing-user-pwd", DbName: pgDefaultUser,
	}

	localState := newFakeStateManager(local_state.State{
		ClusterState: local_state.ClusterState{
			PgRootDsn: rootPg.ConnectionString(),
			PgNodeDsn: nodePg.ConnectionString(),
		},
	})
	payload := &velez_api.EnableStatefullTaskPayload{}

	j := &generateCredentialsJob{localStateManager: localState, ctx: payload}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if payload.GetRootPwd() != "existing-root-pwd" {
		t.Errorf("expected reused root password, got %q", payload.GetRootPwd())
	}

	if payload.GetUserPwd() != "existing-user-pwd" {
		t.Errorf("expected reused user password, got %q", payload.GetUserPwd())
	}
}

func TestGenerateCredentialsJob_AlreadyPersisted_DoesNotRegenerate(t *testing.T) {
	localState := newFakeStateManager(local_state.State{})
	payload := &velez_api.EnableStatefullTaskPayload{
		RootPwd: proto("already-set-root"),
		UserPwd: proto("already-set-user"),
	}

	j := &generateCredentialsJob{localStateManager: localState, ctx: payload}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if payload.GetRootPwd() != "already-set-root" {
		t.Errorf("expected persisted root password preserved, got %q", payload.GetRootPwd())
	}

	if payload.GetUserPwd() != "already-set-user" {
		t.Errorf("expected persisted user password preserved, got %q", payload.GetUserPwd())
	}
}

// createPgContainerJob

func TestCreatePgContainerJob_Success(t *testing.T) {
	docker := newFakeDocker()

	docker.containerCreateResp = container.CreateResponse{ID: testPgContainerID}

	nodeClients := newFakeNodeClients(docker)

	// IsExposePort:true keeps this test's outcome independent of whether the
	// test process itself happens to be running in a container - see the
	// getRootDsnJob tests' comment on env.IsInContainer() for why.
	payload := &velez_api.EnableStatefullTaskPayload{
		Request: &velez_api.EnableStatefullCluster{IsExposePort: protoBool(true)},
		RootPwd: proto("root-pwd"),
	}

	j := &createPgContainerJob{nodeClients: nodeClients, req: payload, pwd: payload, ctx: payload}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if payload.GetContainerId() != testPgContainerID {
		t.Errorf("expected container id 'pg123', got %q", payload.GetContainerId())
	}
}

func TestCreatePgContainerJob_ContainerCreateError(t *testing.T) {
	docker := newFakeDocker()

	docker.containerCreateErr = errNoSpaceLeftOnDevice

	nodeClients := newFakeNodeClients(docker)

	payload := &velez_api.EnableStatefullTaskPayload{
		Request: &velez_api.EnableStatefullCluster{IsExposePort: protoBool(true)},
		RootPwd: proto("root-pwd"),
	}

	j := &createPgContainerJob{nodeClients: nodeClients, req: payload, pwd: payload, ctx: payload}

	err := j.Do(context.Background())
	if err == nil {
		t.Fatal("expected an error when ContainerCreate fails")
	}

	if payload.GetContainerId() != "" {
		t.Errorf("expected no container id set on failure, got %q", payload.GetContainerId())
	}
}

func TestCreatePgContainerJob_BinaryModeWithoutExposePort_Error(t *testing.T) {
	if env.IsInContainer() {
		t.Skip("this failure branch only triggers when velez is not itself running inside a container")
	}

	docker := newFakeDocker()

	docker.containerCreateResp = container.CreateResponse{ID: testPgContainerID}

	nodeClients := newFakeNodeClients(docker)

	payload := &velez_api.EnableStatefullTaskPayload{
		Request: &velez_api.EnableStatefullCluster{IsExposePort: protoBool(false)},
		RootPwd: proto("root-pwd"),
	}

	j := &createPgContainerJob{nodeClients: nodeClients, req: payload, pwd: payload, ctx: payload}

	err := j.Do(context.Background())
	if !errors.Is(err, user_errors.ErrPortMustBeExposedForBinary) {
		t.Fatalf("expected ErrPortMustBeExposedForBinary, got: %v", err)
	}

	if payload.GetContainerId() != "" {
		t.Errorf("expected no container to be created, got id %q", payload.GetContainerId())
	}
}

func TestCreatePgContainerJob_ExposeToPortOccupied_Error(t *testing.T) {
	const requestedPort = 15432

	docker := newFakeDocker()

	docker.containerCreateResp = container.CreateResponse{ID: testPgContainerID}
	docker.listOccupiedPortsResp = []uint32{requestedPort}

	nodeClients := newFakeNodeClients(docker)

	payload := &velez_api.EnableStatefullTaskPayload{
		Request: &velez_api.EnableStatefullCluster{
			IsExposePort: protoBool(true),
			ExposeToPort: protoUint64(requestedPort),
		},
		RootPwd: proto("root-pwd"),
	}

	j := &createPgContainerJob{nodeClients: nodeClients, req: payload, pwd: payload, ctx: payload}

	err := j.Do(context.Background())
	if err == nil {
		t.Fatal("expected an error when the requested port is already occupied")
	}

	if payload.GetContainerId() != "" {
		t.Errorf("expected no container to be created for an occupied port, got id %q", payload.GetContainerId())
	}
}

func TestCreatePgContainerJob_Rollback_NoContainerId_NoOp(t *testing.T) {
	docker := newFakeDocker()
	nodeClients := newFakeNodeClients(docker)
	payload := &velez_api.EnableStatefullTaskPayload{}

	j := &createPgContainerJob{nodeClients: nodeClients, ctx: payload}

	err := j.Rollback(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(docker.removeCalledWith) != 0 {
		t.Errorf("expected Remove not to be called, got %v", docker.removeCalledWith)
	}
}

func TestCreatePgContainerJob_Rollback_RemovesContainer(t *testing.T) {
	docker := newFakeDocker()
	nodeClients := newFakeNodeClients(docker)
	payload := &velez_api.EnableStatefullTaskPayload{ContainerId: proto(testPgContainerID)}

	j := &createPgContainerJob{nodeClients: nodeClients, ctx: payload}

	err := j.Rollback(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(docker.removeCalledWith) != 1 || docker.removeCalledWith[0] != testPgContainerID {
		t.Errorf("expected Remove called with 'pg123', got %v", docker.removeCalledWith)
	}
}

func TestCreatePgContainerJob_Rollback_NotFoundIsSwallowed(t *testing.T) {
	docker := newFakeDocker()

	docker.removeErr = errdefs.NotFound(user_errors.ErrNoSuchContainer)

	nodeClients := newFakeNodeClients(docker)
	payload := &velez_api.EnableStatefullTaskPayload{ContainerId: proto(testPgContainerID)}

	j := &createPgContainerJob{nodeClients: nodeClients, ctx: payload}

	err := j.Rollback(context.Background())
	if err != nil {
		t.Fatalf("expected a not-found Remove error to be swallowed, got %v", err)
	}
}

// startPgContainerJob

func TestStartPgContainerJob_Success(t *testing.T) {
	api := newFakeContainerAPI()
	payload := &velez_api.EnableStatefullTaskPayload{ContainerId: proto(testPgContainerID)}

	j := &startPgContainerJob{dockerAPI: api, ctx: payload}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(api.startCalledWith) != 1 || api.startCalledWith[0] != testPgContainerID {
		t.Errorf("expected container started with 'pg123', got %v", api.startCalledWith)
	}
}

func TestStartPgContainerJob_NoContainerId_Error(t *testing.T) {
	api := newFakeContainerAPI()
	payload := &velez_api.EnableStatefullTaskPayload{}

	j := &startPgContainerJob{dockerAPI: api, ctx: payload}

	err := j.Do(context.Background())
	if err == nil {
		t.Fatal("expected an error when no container id is set")
	}
}

func TestStartPgContainerJob_Rollback_StopsContainer(t *testing.T) {
	api := newFakeContainerAPI()
	payload := &velez_api.EnableStatefullTaskPayload{ContainerId: proto(testPgContainerID)}

	j := &startPgContainerJob{dockerAPI: api, ctx: payload}

	err := j.Rollback(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(api.stopCalledWith) != 1 || api.stopCalledWith[0] != testPgContainerID {
		t.Errorf("expected container stopped with 'pg123', got %v", api.stopCalledWith)
	}
}

// waitForPgReadyJob

func TestWaitForPgReadyJob_AlreadyHealthy_Success(t *testing.T) {
	api := newFakeContainerAPI()

	api.inspectResp = container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{
			State: &container.State{Health: &container.Health{Status: container.Healthy}},
		},
	}

	payload := &velez_api.EnableStatefullTaskPayload{ContainerId: proto(testPgContainerID)}

	j := &waitForPgReadyJob{dockerAPI: api, ctx: payload}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWaitForPgReadyJob_NoContainerId_Error(t *testing.T) {
	api := newFakeContainerAPI()
	payload := &velez_api.EnableStatefullTaskPayload{}

	j := &waitForPgReadyJob{dockerAPI: api, ctx: payload}

	err := j.Do(context.Background())
	if err == nil {
		t.Fatal("expected an error when no container id is set")
	}
}

func TestWaitForPgReadyJob_InspectError(t *testing.T) {
	api := newFakeContainerAPI()

	api.inspectErr = user_errors.ErrNoSuchContainer

	payload := &velez_api.EnableStatefullTaskPayload{ContainerId: proto(testPgContainerID)}

	j := &waitForPgReadyJob{dockerAPI: api, ctx: payload}

	err := j.Do(context.Background())
	if err == nil {
		t.Fatal("expected an error when ContainerInspect fails")
	}
}

func TestWaitForPgReadyJob_NotYetHealthy_ContextCancelled(t *testing.T) {
	api := newFakeContainerAPI()

	api.inspectResp = container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{
			State: &container.State{Health: &container.Health{Status: container.Starting}},
		},
	}

	payload := &velez_api.EnableStatefullTaskPayload{ContainerId: proto(testPgContainerID)}

	j := &waitForPgReadyJob{dockerAPI: api, ctx: payload}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := j.Do(ctx)
	if err == nil {
		t.Fatal("expected an error when the context is cancelled before the container becomes healthy")
	}
}

func TestWaitForPgReadyJob_NilContainerState_DoesNotPanic_ContextCancelled(t *testing.T) {
	api := newFakeContainerAPI()

	api.inspectResp = container.InspectResponse{}

	payload := &velez_api.EnableStatefullTaskPayload{ContainerId: proto(testPgContainerID)}

	j := &waitForPgReadyJob{dockerAPI: api, ctx: payload}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := j.Do(ctx)
	if err == nil {
		t.Fatal("expected an error when the context is cancelled before the container becomes healthy")
	}
}

// getRootDsnJob
//
// env.IsInContainer() memoizes a real filesystem check (/etc/hostname
// readable) for the process lifetime, so these tests avoid asserting on
// that branch directly and instead pick inputs that behave the same way
// regardless of which branch the test process happens to be running under.

func TestGetRootDsnJob_Success_ParsesEnvVars(t *testing.T) {
	api := newFakeContainerAPI()

	networkSettings := &container.NetworkSettings{}

	networkSettings.Ports = nat.PortMap{
		"5432/tcp": []nat.PortBinding{{HostPort: "15432"}},
	}

	api.inspectResp = container.InspectResponse{
		Config: &container.Config{
			Env: []string{
				"POSTGRES_DB=postgres",
				"POSTGRES_USER=postgres",
				"POSTGRES_PASSWORD=root-pwd",
			},
		},
		NetworkSettings: networkSettings,
	}

	payload := &velez_api.EnableStatefullTaskPayload{
		Request:     &velez_api.EnableStatefullCluster{IsExposePort: protoBool(true)},
		ContainerId: proto(testPgContainerID),
	}

	j := &getRootDsnJob{dockerAPI: api, ctx: payload}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got resources.Postgres

	err = got.ParseFromDsn(payload.GetRootDsn())
	if err != nil {
		t.Fatalf("expected a parseable dsn, got error: %v", err)
	}

	if got.User != pgDefaultUser || got.Pwd != "root-pwd" || got.DbName != pgDefaultUser {
		t.Errorf("expected user/pwd/db to be parsed from container env, got %+v", got)
	}
}

func TestGetRootDsnJob_InspectError(t *testing.T) {
	api := newFakeContainerAPI()

	api.inspectErr = user_errors.ErrNoSuchContainer

	payload := &velez_api.EnableStatefullTaskPayload{ContainerId: proto(testPgContainerID)}

	j := &getRootDsnJob{dockerAPI: api, ctx: payload}

	err := j.Do(context.Background())
	if err == nil {
		t.Fatal("expected an error when ContainerInspect fails")
	}
}

// createSchemaAndMigrateJob / createPgUserJob
//
// Both jobs reuse sqldb.RollMigration/cluster_steps.CreatePgUserForNode
// verbatim, which open a real Postgres connection - same as the original,
// never-unit-tested pipeline steps (see docs/jobs_migrations/questions.md).
// Only the connection-failure path is exercised here; the success path has
// no unit-level coverage and needs a real Postgres (covered by tests/e2e).

func TestCreateSchemaAndMigrateJob_ConnectionFailure(t *testing.T) {
	payload := &velez_api.EnableStatefullTaskPayload{
		RootDsn: proto("postgres://u:p@127.0.0.1:1/db?sslmode=disable"),
	}

	j := &createSchemaAndMigrateJob{dsn: payload}

	err := j.Do(context.Background())
	if err == nil {
		t.Fatal("expected an error connecting to an unreachable postgres port")
	}
}

func TestCreatePgUserJob_ConnectionFailure(t *testing.T) {
	payload := &velez_api.EnableStatefullTaskPayload{
		RootDsn: proto("postgres://u:p@127.0.0.1:1/db?sslmode=disable"),
		UserPwd: proto("user-pwd"),
	}

	j := &createPgUserJob{dsn: payload, pwd: payload}

	err := j.Do(context.Background())
	if err == nil {
		t.Fatal("expected an error connecting to an unreachable postgres port")
	}
}

// updateClusterStateJob

func TestUpdateClusterStateJob_Success(t *testing.T) {
	localStateManager := newFakeStateManager(local_state.State{})

	initialStorage := &fakeClusterStorage{nodes: &fakeNodesStorage{}}
	clusterStateManager := state.NewContainer(initialStorage)
	storageContainer := storage.NewStorageContainer(initialStorage)

	newStorage := &fakeClusterStorage{nodes: &fakeNodesStorage{}}

	var gotDsn string

	newPgStateManager := func(_ context.Context, dsn string) (cluster_clients.ClusterStateManager, error) {
		gotDsn = dsn

		return newStorage, nil
	}

	payload := &velez_api.EnableStatefullTaskPayload{
		RootDsn: proto("postgresql://postgres:root-pwd@h:5432/postgres?sslmode=disable"),
		UserPwd: proto("user-pwd"),
	}

	j := &updateClusterStateJob{
		localStateManager:   localStateManager,
		clusterStateManager: clusterStateManager,
		storageContainer:    storageContainer,
		newPgStateManager:   newPgStateManager,
		dsn:                 payload,
		pwd:                 payload,
	}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotDsn == "" {
		t.Error("expected newPgStateManager to be called with a node dsn")
	}

	finalState := localStateManager.Get()
	if finalState.ClusterState.PgRootDsn != payload.GetRootDsn() {
		t.Errorf("expected PgRootDsn to be persisted, got %q", finalState.ClusterState.PgRootDsn)
	}

	if finalState.ClusterState.PgNodeDsn == "" {
		t.Error("expected PgNodeDsn to be persisted")
	}

	if clusterStateManager.Nodes() != newStorage.nodes {
		t.Error("expected clusterStateManager to be swapped to the new storage")
	}

	if storageContainer.Nodes() != newStorage.nodes {
		t.Error("expected storageContainer to be swapped to the new storage")
	}
}

func TestUpdateClusterStateJob_NewPgStateManagerError(t *testing.T) {
	localStateManager := newFakeStateManager(local_state.State{})

	initialStorage := &fakeClusterStorage{nodes: &fakeNodesStorage{}}
	clusterStateManager := state.NewContainer(initialStorage)
	storageContainer := storage.NewStorageContainer(initialStorage)

	newPgStateManager := func(_ context.Context, _ string) (cluster_clients.ClusterStateManager, error) {
		return nil, errConnectionRefused
	}

	payload := &velez_api.EnableStatefullTaskPayload{
		RootDsn: proto("postgresql://postgres:root-pwd@h:5432/postgres?sslmode=disable"),
		UserPwd: proto("user-pwd"),
	}

	j := &updateClusterStateJob{
		localStateManager:   localStateManager,
		clusterStateManager: clusterStateManager,
		storageContainer:    storageContainer,
		newPgStateManager:   newPgStateManager,
		dsn:                 payload,
		pwd:                 payload,
	}

	err := j.Do(context.Background())
	if err == nil {
		t.Fatal("expected an error when newPgStateManager fails")
	}

	if clusterStateManager.Nodes() != initialStorage.nodes {
		t.Error("expected clusterStateManager to remain unchanged on failure")
	}
}

// initNodeStorageJob

func TestInitNodeStorageJob_Success(t *testing.T) {
	nodes := &fakeNodesStorage{}
	clusterStateManager := state.NewContainer(&fakeClusterStorage{nodes: nodes})

	j := &initNodeStorageJob{clusterStateManager: clusterStateManager, region: "eu-west"}

	err := j.Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if nodes.initNodeCalls != 1 {
		t.Errorf("expected InitNode called once, got %d", nodes.initNodeCalls)
	}

	if nodes.lastRegion != "eu-west" {
		t.Errorf("expected region %q passed through to InitNode, got %q", "eu-west", nodes.lastRegion)
	}
}

func TestInitNodeStorageJob_Error(t *testing.T) {
	nodes := &fakeNodesStorage{initNodeErr: errDbUnreachable}
	clusterStateManager := state.NewContainer(&fakeClusterStorage{nodes: nodes})

	j := &initNodeStorageJob{clusterStateManager: clusterStateManager}

	err := j.Do(context.Background())
	if err == nil {
		t.Fatal("expected an error when InitNode fails")
	}
}

// End-to-end coverage through the real handler + taskWorker, following
// connect_service_to_vpn_test.go's patch-field-after-BuildJobs precedent:
// startPgContainerJob's/getRootDsnJob's docker fields are
// nodeClients.Docker().Client() (nil on fakeDocker), so they're swapped for a
// fakeContainerAPI after BuildJobs before running.
//
// This only exercises the first 4 jobs successfully: create_schema_and_migrate
// needs a real Postgres connection (see the createSchemaAndMigrateJob comment
// above), so it's left to fail against an unreachable dsn here - which
// doubles as this handler's required failure-path test, exercising rollback
// of start_container/create_container through the real checkpoint chain.

func enableStatefullTask(
	t *testing.T, tasksStorage *fakeTasksStorage, entityID string, payload *velez_api.EnableStatefullTaskPayload,
) tasks_queries.VelezTask {
	t.Helper()

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("unexpected error marshaling payload: %v", err)
	}

	taskCtx := pqtype.NullRawMessage{RawMessage: payloadJSON, Valid: true}
	params := tasks_queries.CreateTaskParams{
		EntityID: entityID,
		Action:   EnableStatefullAction,
		Context:  taskCtx,
	}

	task, err := tasksStorage.CreateTask(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error creating task: %v", err)
	}

	return task
}

func patchEnableStatefullDockerFields(namedJobs []NamedJob, containerAPI *fakeContainerAPI) {
	for _, nj := range namedJobs {
		switch j := nj.Job.(type) {
		case *startPgContainerJob:
			j.dockerAPI = containerAPI
		case *waitForPgReadyJob:
			j.dockerAPI = containerAPI
		case *getRootDsnJob:
			j.dockerAPI = containerAPI
		}
	}
}

func TestEnableStatefullHandler_FailurePath_UnreachablePostgres_RollsBack(t *testing.T) {
	tasksStorage := newFakeTasksStorage()
	jobsStorage := newFakeJobsStorage()

	request := &velez_api.EnableStatefullCluster{IsExposePort: protoBool(true)}
	payload := &velez_api.EnableStatefullTaskPayload{Request: request}
	task := enableStatefullTask(t, tasksStorage, "cluster", payload)

	docker := newFakeDocker()

	docker.containerCreateResp = container.CreateResponse{ID: testPgContainerID}

	nodeClients := newFakeNodeClients(docker)

	nodeClients.localState = newFakeStateManager(local_state.State{})

	clusterStorage := &fakeClusterStorage{nodes: &fakeNodesStorage{}}
	clusterStateManager := state.NewContainer(clusterStorage)
	storageContainer := storage.NewStorageContainer(clusterStorage)

	handler := NewEnableStatefullHandler(nodeClients, clusterStateManager, storageContainer, config.Config{})

	taskCtx := handler.NewContext()

	err := json.Unmarshal(task.Context.RawMessage, taskCtx)
	if err != nil {
		t.Fatalf("unexpected error unmarshaling task context: %v", err)
	}

	namedJobs := handler.BuildJobs(taskCtx)

	containerAPI := newFakeContainerAPI()

	networkSettings := &container.NetworkSettings{}

	networkSettings.Ports = nat.PortMap{"5432/tcp": []nat.PortBinding{{HostPort: "1"}}}

	containerAPI.inspectResp = container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{
			State: &container.State{Health: &container.Health{Status: container.Healthy}},
		},
		Config: &container.Config{
			Env: []string{"POSTGRES_DB=postgres", "POSTGRES_USER=postgres", "POSTGRES_PASSWORD=root-pwd"},
		},
		NetworkSettings: networkSettings,
	}
	patchEnableStatefullDockerFields(namedJobs, containerAPI)

	registry := NewRegistry()
	registry.Register(handler)

	w, ok := NewTaskWorker(tasksStorage, jobsStorage, registry, "test-worker", time.Hour).(*taskWorker)
	if !ok {
		t.Fatal("expected NewTaskWorker to return *taskWorker")
	}

	// w.run rebuilds its own namedJobs internally via handler.BuildJobs, which
	// would discard the docker-field patch above - so this drives runJobs
	// directly (same as connect_service_to_vpn_test.go's happy-path test) and
	// replicates w.run's FinishTask bookkeeping by hand.
	runErr := w.runJobs(context.Background(), task.ID, taskCtx, namedJobs)
	if runErr == nil {
		t.Fatal("expected an error: create_schema_and_migrate can't reach a real postgres in this test")
	}

	err = tasksStorage.FinishTask(context.Background(), tasks_queries.FinishTaskParams{
		ID:     task.ID,
		Status: tasks_queries.VelezTaskStatusFAILED,
		Error:  sql.NullString{String: runErr.Error(), Valid: true},
	})
	if err != nil {
		t.Fatalf("unexpected error finishing task: %v", err)
	}

	finished := tasksStorage.get(task.ID)
	if finished.Status != tasks_queries.VelezTaskStatusFAILED {
		t.Errorf("expected task status FAILED, got %v", finished.Status)
	}

	stepNames := []string{
		stepGenerateCredentials, stepCreatePgContainer, stepStartSidecar, "wait_for_postgres_ready", stepGetRootDsn,
	}
	for _, name := range stepNames {
		row, ok := jobsStorage.rows[jobKey(task.ID, name)]
		if !ok || row.Status != jobs_queries.VelezJobStatusDONE {
			t.Errorf("expected job %q checkpointed DONE, got %+v (ok=%v)", name, row, ok)
		}
	}

	if len(containerAPI.stopCalledWith) != 1 || containerAPI.stopCalledWith[0] != testPgContainerID {
		t.Errorf("expected the postgres container stopped on rollback, got %v", containerAPI.stopCalledWith)
	}

	if len(docker.removeCalledWith) != 1 || docker.removeCalledWith[0] != testPgContainerID {
		t.Errorf("expected the postgres container removed on rollback, got %v", docker.removeCalledWith)
	}
}

func protoBool(b bool) *bool {
	return &b
}

func protoUint64(v uint64) *uint64 {
	return &v
}
