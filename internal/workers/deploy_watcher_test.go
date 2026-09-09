package workers

import (
	"context"
	"database/sql"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/sqlc-dev/pqtype"
	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/deployments_queries"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
	"go.vervstack.ru/Velez/tests/test_helper"
)

const (
	testDeploymentID  = int64(7)
	testSpecID        = int64(42)
	testSmerdName     = "hello_world"
	testSmerdImage    = "godverv/hello_world:v0.0.14"
	testEnvironment   = "staging"
	testTaskFailedMsg = "container create blew up"
)

// enqueuedTask records one taskRunner.Enqueue call.
type enqueuedTask struct {
	entityID string
	action   string
	// contextJSON is the initial context marshaled exactly the way
	// jobs.Engine.Enqueue persists it, so assertions see what a real worker
	// would later unmarshal.
	contextJSON []byte
}

// fakeTaskRunner is a pure in-process double for the jobs engine: it records
// what deployWatcher enqueued and immediately reports the task terminal, so
// the worker's dispatch logic can be exercised without a task worker, a
// database or Docker. No external dependency is faked - per CLAUDE.md this is
// the "records call order" category of test double, not a mocked API.
type fakeTaskRunner struct {
	mu   sync.Mutex
	seen []enqueuedTask

	// finalStatus is the status Watch reports for every task.
	finalStatus tasks_queries.VelezTaskStatus
	finalError  sql.NullString
	// enqueueErr, when set, makes Enqueue fail without recording anything.
	enqueueErr error
	// blockWatch makes Watch never emit, so the caller's watch timeout wins.
	blockWatch bool
}

func newFakeTaskRunner() *fakeTaskRunner {
	return &fakeTaskRunner{
		finalStatus: tasks_queries.VelezTaskStatusDONE,
	}
}

func (f *fakeTaskRunner) Enqueue(
	_ context.Context, entityID, action string, initialContext any,
) (tasks_queries.VelezTask, error) {
	if f.enqueueErr != nil {
		return tasks_queries.VelezTask{}, f.enqueueErr
	}

	contextJSON, err := json.Marshal(initialContext)
	if err != nil {
		return tasks_queries.VelezTask{}, rerrors.Wrap(err, "error marshaling initial context")
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	task := enqueuedTask{
		entityID:    entityID,
		action:      action,
		contextJSON: contextJSON,
	}

	f.seen = append(f.seen, task)

	return tasks_queries.VelezTask{EntityID: entityID, Action: action}, nil
}

func (f *fakeTaskRunner) Watch(ctx context.Context, entityID, action string) <-chan tasks_queries.VelezTask {
	ch := make(chan tasks_queries.VelezTask, 1)

	if f.blockWatch {
		go func() {
			<-ctx.Done()
			close(ch)
		}()

		return ch
	}

	task := tasks_queries.VelezTask{
		EntityID: entityID,
		Action:   action,
		Status:   f.finalStatus,
		Error:    f.finalError,
	}

	ch <- task

	close(ch)

	return ch
}

func (f *fakeTaskRunner) calls() []enqueuedTask {
	f.mu.Lock()
	defer f.mu.Unlock()

	out := make([]enqueuedTask, len(f.seen))
	copy(out, f.seen)

	return out
}

// stubDeploymentsStorage serves one stored specification and records every
// status transition the watcher writes.
type stubDeploymentsStorage struct {
	mu sync.Mutex

	spec     deployments_queries.GetSpecificationByIdRow
	statuses []deployments_queries.UpdateDeploymentStatusParams
}

func newStubDeploymentsStorage(t *testing.T, req *velez_api.CreateSmerd_Request) *stubDeploymentsStorage {
	t.Helper()

	payload, err := json.Marshal(req)
	require.NoError(t, err)

	spec := deployments_queries.GetSpecificationByIdRow{
		ID:          testSpecID,
		Name:        req.GetName(),
		VervPayload: pqtype.NullRawMessage{RawMessage: payload, Valid: true},
	}

	return &stubDeploymentsStorage{spec: spec}
}

func (s *stubDeploymentsStorage) GetSpecificationById(
	_ context.Context, _ int64,
) (deployments_queries.GetSpecificationByIdRow, error) {
	return s.spec, nil
}

func (s *stubDeploymentsStorage) UpdateDeploymentStatus(
	_ context.Context, arg deployments_queries.UpdateDeploymentStatusParams,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.statuses = append(s.statuses, arg)

	return nil
}

func (s *stubDeploymentsStorage) CreateSpecification(
	_ context.Context, _ deployments_queries.CreateSpecificationParams,
) (int64, error) {
	return 0, nil
}

func (s *stubDeploymentsStorage) CreateDeployment(
	_ context.Context, _ deployments_queries.CreateDeploymentParams,
) (any, error) {
	return nil, nil
}

func (s *stubDeploymentsStorage) List(
	_ context.Context, _ domain.ListDeploymentsReq,
) ([]domain.Deployment, error) {
	return nil, nil
}

func (s *stubDeploymentsStorage) ListDeployments(
	_ context.Context, _ domain.ListDeploymentsReq,
) (domain.DeploymentList, error) {
	return domain.DeploymentList{}, nil
}

func (s *stubDeploymentsStorage) WithTx(_ *sql.Tx) deployments_queries.Querier {
	return s
}

func (s *stubDeploymentsStorage) recordedStatuses() []deployments_queries.UpdateDeploymentStatusParams {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]deployments_queries.UpdateDeploymentStatusParams, len(s.statuses))
	copy(out, s.statuses)

	return out
}

func testCreateRequest() *velez_api.CreateSmerd_Request {
	return &velez_api.CreateSmerd_Request{
		Name:        testSmerdName,
		ImageName:   testSmerdImage,
		Environment: testEnvironment,
	}
}

// stubStorageResolver re-serves the same deployments storage per call, standing
// in for the swappable cluster state manager the real watcher now holds.
type stubStorageResolver struct{ d storage.DeploymentsStorage }

func (s stubStorageResolver) Deployments() storage.DeploymentsStorage { return s.d }

func newTestWatcher(
	runner *fakeTaskRunner, deployments *stubDeploymentsStorage, runtimes container_runtime.RuntimeResolver,
) *deployWatcher {
	return &deployWatcher{
		jobsEngine:  runner,
		dataStorage: stubStorageResolver{d: deployments},
		runtimes:    runtimes,
		nodeId:      1,
		ticker:      time.NewTicker(time.Hour),
		done:        make(chan struct{}),
	}
}

func scheduledDeployment(status deployments_queries.VelezDeploymentStatus) domain.Deployment {
	return domain.Deployment{
		Id:     testDeploymentID,
		SpecId: testSpecID,
		Status: status,
	}
}

// A scheduled deployment must become a create_smerd task keyed on the smerd's
// name scoped by its environment (jobs.SmerdEntityID), carrying the stored
// request verbatim - including its environment, which the deleted pipeliner
// used to pre-resolve into a Docker suffix here and which create_smerd's own
// jobs now re-resolve at run time.
func TestDeployWatcher_ScheduledDeploymentEnqueuesCreateSmerd(t *testing.T) {
	t.Parallel()

	runner := newFakeTaskRunner()
	deployments := newStubDeploymentsStorage(t, testCreateRequest())
	watcher := newTestWatcher(runner, deployments, nil)

	batch := []domain.Deployment{
		scheduledDeployment(deployments_queries.VelezDeploymentStatusSCHEDULEDDEPLOYMENT),
	}

	err := watcher.processScheduledBatch(t.Context(), batch)
	require.NoError(t, err)

	calls := runner.calls()
	require.Len(t, calls, 1)
	require.Equal(t, jobs.CreateSmerdAction, calls[0].action)
	require.Equal(t, testEnvironment+"/"+testSmerdName, calls[0].entityID)

	payload := &velez_api.CreateSmerdTaskPayload{}

	err = json.Unmarshal(calls[0].contextJSON, payload)
	require.NoError(t, err)

	require.Equal(t, testSmerdName, payload.GetRequest().GetName())
	require.Equal(t, testSmerdImage, payload.GetRequest().GetImageName())
	require.Equal(t, testEnvironment, payload.GetRequest().GetEnvironment())

	statuses := deployments.recordedStatuses()
	require.Len(t, statuses, 1)
	require.Equal(t, deployments_queries.VelezDeploymentStatusRUNNING, statuses[0].Status)
	require.Equal(t, testDeploymentID, statuses[0].ID)
}

// A scheduled upgrade must become an upgrade_smerd task that carries the
// specification's environment. The pipeliner path read the environment for
// deployments but dropped it for upgrades (domain.UpgradeSmerd had no such
// field), so every upgrade silently targeted the default environment.
func TestDeployWatcher_ScheduledUpgradeEnqueuesUpgradeSmerdWithEnvironment(t *testing.T) {
	t.Parallel()

	runner := newFakeTaskRunner()
	deployments := newStubDeploymentsStorage(t, testCreateRequest())
	watcher := newTestWatcher(runner, deployments, nil)

	batch := []domain.Deployment{
		scheduledDeployment(deployments_queries.VelezDeploymentStatusSCHEDULEDUPGRADE),
	}

	err := watcher.processScheduledBatch(t.Context(), batch)
	require.NoError(t, err)

	calls := runner.calls()
	require.Len(t, calls, 1)
	require.Equal(t, jobs.UpgradeSmerdAction, calls[0].action)
	require.Equal(t, testEnvironment+"/"+testSmerdName, calls[0].entityID)

	payload := &velez_api.UpgradeSmerdTaskPayload{}

	err = json.Unmarshal(calls[0].contextJSON, payload)
	require.NoError(t, err)

	require.Equal(t, testSmerdName, payload.GetUpgradeRequest().GetName())
	require.Equal(t, testSmerdImage, payload.GetUpgradeRequest().GetImage())
	require.Equal(t, testEnvironment, payload.GetUpgradeRequest().GetEnvironment())

	statuses := deployments.recordedStatuses()
	require.Len(t, statuses, 1)
	require.Equal(t, deployments_queries.VelezDeploymentStatusRUNNING, statuses[0].Status)
}

// A FAILED task must mark the deployment FAILED rather than RUNNING - the
// contract the pipeliner's `runner.Run(ctx)` error used to carry.
func TestDeployWatcher_FailedTaskMarksDeploymentFailed(t *testing.T) {
	t.Parallel()

	runner := newFakeTaskRunner()

	runner.finalStatus = tasks_queries.VelezTaskStatusFAILED
	runner.finalError = sql.NullString{String: testTaskFailedMsg, Valid: true}

	deployments := newStubDeploymentsStorage(t, testCreateRequest())
	watcher := newTestWatcher(runner, deployments, nil)

	batch := []domain.Deployment{
		scheduledDeployment(deployments_queries.VelezDeploymentStatusSCHEDULEDDEPLOYMENT),
	}

	err := watcher.processScheduledBatch(t.Context(), batch)
	require.NoError(t, err)

	statuses := deployments.recordedStatuses()
	require.Len(t, statuses, 1)
	require.Equal(t, deployments_queries.VelezDeploymentStatusFAILED, statuses[0].Status)
}

// runTask must surface the task's own error message, not swallow it.
func TestDeployWatcher_RunTaskReturnsTaskError(t *testing.T) {
	t.Parallel()

	runner := newFakeTaskRunner()

	runner.finalStatus = tasks_queries.VelezTaskStatusFAILED
	runner.finalError = sql.NullString{String: testTaskFailedMsg, Valid: true}

	watcher := newTestWatcher(runner, newStubDeploymentsStorage(t, testCreateRequest()), nil)

	err := watcher.runTask(t.Context(), testSmerdName, jobs.CreateSmerdAction, testCreateRequest())
	require.ErrorContains(t, err, testTaskFailedMsg)
}

// A task that never reaches a terminal status must time out instead of
// blocking the watcher forever.
func TestDeployWatcher_RunTaskTimesOut(t *testing.T) {
	t.Parallel()

	runner := newFakeTaskRunner()

	runner.blockWatch = true

	watcher := newTestWatcher(runner, newStubDeploymentsStorage(t, testCreateRequest()), nil)

	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()

	err := watcher.runTask(ctx, testSmerdName, jobs.CreateSmerdAction, testCreateRequest())
	require.ErrorContains(t, err, "timed out waiting for")
}

// deleteBatch removes the container through the resolved ContainerRuntime
// instead of the raw Docker client it used before, then marks the deployment
// DELETED. Runs against a real Docker daemon per this repo's testing rules.
func TestDeployWatcher_DeleteBatchRemovesContainerThroughRuntime(t *testing.T) {
	t.Parallel()

	api := test_helper.NewRealDockerAPI(t)
	test_helper.EnsurePulled(t, api, test_helper.HelloWorldAppImage)

	name := test_helper.UniqueName(t, "velez_dw")

	cfg := &container.Config{
		Image: test_helper.HelloWorldAppImage,
	}

	created, err := api.ContainerCreate(t.Context(), cfg, nil, nil, nil, name)
	require.NoError(t, err)

	t.Cleanup(func() { test_helper.RemoveContainer(t, api, created.ID) })

	req := &velez_api.CreateSmerd_Request{
		Name:      name,
		ImageName: test_helper.HelloWorldAppImage,
	}

	runner := newFakeTaskRunner()
	deployments := newStubDeploymentsStorage(t, req)
	runtimes := container_runtime.NewResolver(api, nil, nil)
	watcher := newTestWatcher(runner, deployments, runtimes)

	batch := []domain.Deployment{
		scheduledDeployment(deployments_queries.VelezDeploymentStatusSCHEDULEDDELETION),
	}

	err = watcher.deleteBatch(t.Context(), batch)
	require.NoError(t, err)

	_, err = api.ContainerInspect(t.Context(), created.ID)
	require.Error(t, err)

	statuses := deployments.recordedStatuses()
	require.Len(t, statuses, 1)
	require.Equal(t, deployments_queries.VelezDeploymentStatusDELETED, statuses[0].Status)
}
