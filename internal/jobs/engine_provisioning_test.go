package jobs

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/local_storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/jobs_queries"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	provisioningActionA = "provisioning_action_a"
	provisioningActionB = "provisioning_action_b"
	pendingEntity       = "pending"
)

type provisioningFixture struct {
	engine       Engine
	tasksStorage storage.TasksStorage
	jobsStorage  storage.JobsStorage
}

func newProvisioningFixture() provisioningFixture {
	tasksStorage := local_storage.NewTasksStorage()
	jobsStorage := local_storage.NewJobsStorage()

	registry := NewRegistry()
	registry.Register(&testHandler{
		action: provisioningActionA,
		jobs:   twoNamedJobs,
	})

	engine := NewEngine(tasksStorage, jobsStorage)
	engine.SetRegistry(registry)

	return provisioningFixture{
		engine:       engine,
		tasksStorage: tasksStorage,
		jobsStorage:  jobsStorage,
	}
}

func twoNamedJobs(TaskContext) []NamedJob {
	return []NamedJob{
		{Name: testJobNameFirst, Job: &recordingJob{}},
		{Name: testJobNameSecond, Job: &recordingJob{}},
	}
}

func (f provisioningFixture) enqueue(t *testing.T, entityID, action string) tasks_queries.VelezTask {
	t.Helper()

	task, err := f.engine.Enqueue(context.Background(), entityID, action, &dummyContext{Value: entityID})
	require.NoError(t, err)

	return task
}

func (f provisioningFixture) finish(t *testing.T, taskID int64, status tasks_queries.VelezTaskStatus) {
	t.Helper()

	params := tasks_queries.FinishTaskParams{
		Status: status,
		Error:  sql.NullString{String: "boom", Valid: status == tasks_queries.VelezTaskStatusFAILED},
		ID:     taskID,
	}

	err := f.tasksStorage.FinishTask(context.Background(), params)
	require.NoError(t, err)
}

func (f provisioningFixture) claim(t *testing.T) {
	t.Helper()

	_, err := f.tasksStorage.ClaimTask(context.Background(), tasks_queries.ClaimTaskParams{})
	require.NoError(t, err)
}

func (f provisioningFixture) addJob(
	t *testing.T, taskID int64, name string, status jobs_queries.VelezJobStatus,
) {
	t.Helper()

	createParams := jobs_queries.CreateRunningJobParams{TaskID: taskID, JobName: name}

	_, err := f.jobsStorage.CreateRunningJob(context.Background(), createParams)
	require.NoError(t, err)

	if status == jobs_queries.VelezJobStatusRUNNING {
		return
	}

	finishParams := jobs_queries.FinishJobParams{Status: status, TaskID: taskID, JobName: name}

	err = f.jobsStorage.FinishJob(context.Background(), finishParams)
	require.NoError(t, err)
}

func (f provisioningFixture) jobCount(t *testing.T, taskID int64) int {
	t.Helper()

	rows, err := f.jobsStorage.ListJobsByTask(context.Background(), taskID)
	require.NoError(t, err)

	return len(rows)
}

func entityIDs(entries []ProvisioningEntry) []string {
	ids := make([]string, 0, len(entries))
	for _, entry := range entries {
		ids = append(ids, entry.Task.EntityID)
	}

	return ids
}

func Test_EnqueueReplacing_AbsentTaskIsCreated(t *testing.T) {
	f := newProvisioningFixture()

	task, err := f.engine.EnqueueReplacing(context.Background(), "e1", provisioningActionA, &dummyContext{})
	require.NoError(t, err)

	require.Equal(t, "e1", task.EntityID)
	require.Equal(t, tasks_queries.VelezTaskStatusPENDING, task.Status)
}

func Test_EnqueueReplacing_DoneTaskIsReplacedWithFreshOne(t *testing.T) {
	f := newProvisioningFixture()
	old := f.enqueue(t, "e1", provisioningActionA)
	f.addJob(t, old.ID, testJobNameFirst, jobs_queries.VelezJobStatusDONE)
	f.finish(t, old.ID, tasks_queries.VelezTaskStatusDONE)

	task, err := f.engine.EnqueueReplacing(context.Background(), "e1", provisioningActionA, &dummyContext{})
	require.NoError(t, err)

	require.NotEqual(t, old.ID, task.ID)
	require.Equal(t, tasks_queries.VelezTaskStatusPENDING, task.Status)
	require.Zero(t, f.jobCount(t, old.ID))
}

func Test_EnqueueReplacing_FailedTaskIsReplacedWithFreshOne(t *testing.T) {
	f := newProvisioningFixture()
	old := f.enqueue(t, "e1", provisioningActionA)
	f.addJob(t, old.ID, testJobNameFirst, jobs_queries.VelezJobStatusFAILED)
	f.finish(t, old.ID, tasks_queries.VelezTaskStatusFAILED)

	task, err := f.engine.EnqueueReplacing(context.Background(), "e1", provisioningActionA, &dummyContext{})
	require.NoError(t, err)

	require.NotEqual(t, old.ID, task.ID)
	require.Equal(t, tasks_queries.VelezTaskStatusPENDING, task.Status)
	require.False(t, task.Error.Valid)
	require.Zero(t, f.jobCount(t, old.ID))
}

func Test_EnqueueReplacing_InFlightTaskIsReturnedUnchanged(t *testing.T) {
	f := newProvisioningFixture()
	pending := f.enqueue(t, "e1", provisioningActionA)

	got, err := f.engine.EnqueueReplacing(context.Background(), "e1", provisioningActionA, &dummyContext{})
	require.NoError(t, err)
	require.Equal(t, pending.ID, got.ID)

	f.claim(t)

	got, err = f.engine.EnqueueReplacing(context.Background(), "e1", provisioningActionA, &dummyContext{})
	require.NoError(t, err)
	require.Equal(t, pending.ID, got.ID)
	require.Equal(t, tasks_queries.VelezTaskStatusRUNNING, got.Status)
}

func Test_ListProvisioning_ReturnsInFlightAndRecentlyFailedOnly(t *testing.T) {
	f := newProvisioningFixture()
	f.enqueue(t, "running", provisioningActionA)
	f.claim(t)
	f.enqueue(t, pendingEntity, provisioningActionA)

	failed := f.enqueue(t, "failed", provisioningActionA)
	f.finish(t, failed.ID, tasks_queries.VelezTaskStatusFAILED)

	done := f.enqueue(t, "done", provisioningActionA)
	f.finish(t, done.ID, tasks_queries.VelezTaskStatusDONE)

	entries, err := f.engine.ListProvisioning(
		context.Background(), []string{provisioningActionA}, time.Now().Add(-time.Hour))
	require.NoError(t, err)

	require.ElementsMatch(t, []string{"running", pendingEntity, "failed"}, entityIDs(entries))
}

func Test_ListProvisioning_ExcludesFailedOlderThanWindow(t *testing.T) {
	f := newProvisioningFixture()
	failed := f.enqueue(t, "failed", provisioningActionA)
	f.finish(t, failed.ID, tasks_queries.VelezTaskStatusFAILED)
	f.enqueue(t, pendingEntity, provisioningActionA)

	entries, err := f.engine.ListProvisioning(
		context.Background(), []string{provisioningActionA}, time.Now().Add(time.Hour))
	require.NoError(t, err)

	require.Equal(t, []string{pendingEntity}, entityIDs(entries))
}

func Test_ListProvisioning_FiltersByGivenActions(t *testing.T) {
	f := newProvisioningFixture()
	f.enqueue(t, "a", provisioningActionA)
	f.enqueue(t, "b", provisioningActionB)

	entries, err := f.engine.ListProvisioning(
		context.Background(), []string{provisioningActionB}, time.Now().Add(-time.Hour))
	require.NoError(t, err)

	require.Equal(t, []string{"b"}, entityIDs(entries))
}

func Test_ListProvisioning_AttachesPerJobStatuses(t *testing.T) {
	f := newProvisioningFixture()
	task := f.enqueue(t, "e1", provisioningActionA)
	f.addJob(t, task.ID, testJobNameFirst, jobs_queries.VelezJobStatusDONE)

	entries, err := f.engine.ListProvisioning(
		context.Background(), []string{provisioningActionA}, time.Now().Add(-time.Hour))
	require.NoError(t, err)

	require.Len(t, entries, 1)
	require.Equal(t, []JobStatus{
		{Name: testJobNameFirst, Status: jobs_queries.VelezJobStatusDONE},
		{Name: testJobNameSecond},
	}, entries[0].Jobs)
}

func Test_DismissFailed_DeletesFailedTaskWithJobs(t *testing.T) {
	f := newProvisioningFixture()
	task := f.enqueue(t, "e1", provisioningActionA)
	f.addJob(t, task.ID, testJobNameFirst, jobs_queries.VelezJobStatusFAILED)
	f.finish(t, task.ID, tasks_queries.VelezTaskStatusFAILED)

	err := f.engine.DismissFailed(context.Background(), "e1", provisioningActionA)
	require.NoError(t, err)

	_, err = f.tasksStorage.GetTaskById(context.Background(), task.ID)
	require.ErrorIs(t, err, sql.ErrNoRows)
	require.Zero(t, f.jobCount(t, task.ID))
}

func Test_DismissFailed_RunningTaskIsNotDismissable(t *testing.T) {
	f := newProvisioningFixture()
	task := f.enqueue(t, "e1", provisioningActionA)
	f.claim(t)

	err := f.engine.DismissFailed(context.Background(), "e1", provisioningActionA)
	require.ErrorIs(t, err, user_errors.ErrTaskNotDismissable)

	_, err = f.tasksStorage.GetTaskById(context.Background(), task.ID)
	require.NoError(t, err)
}

func Test_DismissFailed_MissingTaskIsNotFound(t *testing.T) {
	f := newProvisioningFixture()

	err := f.engine.DismissFailed(context.Background(), "e1", provisioningActionA)
	require.ErrorIs(t, err, user_errors.ErrTaskNotFound)
}
