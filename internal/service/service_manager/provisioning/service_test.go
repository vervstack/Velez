package provisioning

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/local_storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/jobs_queries"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	testAction  = "test_provisioning_action"
	firstJob    = "first"
	secondJob   = "second"
	thirdJob    = "third"
	failureText = "boom"
)

type emptyContext struct{}

type threeJobsHandler struct{}

func (threeJobsHandler) Action() string               { return testAction }
func (threeJobsHandler) NewContext() jobs.TaskContext { return &emptyContext{} }
func (threeJobsHandler) BuildJobs(jobs.TaskContext) []jobs.NamedJob {
	return []jobs.NamedJob{
		{Name: firstJob},
		{Name: secondJob},
		{Name: thirdJob},
	}
}

type windowRecordingEngine struct {
	jobs.Engine

	failedSince time.Time
}

func (w *windowRecordingEngine) ListProvisioning(
	ctx context.Context, actions []string, failedSince time.Time,
) ([]jobs.ProvisioningEntry, error) {
	w.failedSince = failedSince

	entries, err := w.Engine.ListProvisioning(ctx, actions, failedSince)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing provisioning")
	}

	return entries, nil
}

type fixture struct {
	service      *Service
	engine       *windowRecordingEngine
	tasksStorage storage.TasksStorage
	jobsStorage  storage.JobsStorage
}

func newFixture() fixture {
	tasksStorage := local_storage.NewTasksStorage()
	jobsStorage := local_storage.NewJobsStorage()

	registry := jobs.NewRegistry()
	registry.Register(threeJobsHandler{})

	realEngine := jobs.NewEngine(tasksStorage, jobsStorage)
	realEngine.SetRegistry(registry)

	engine := &windowRecordingEngine{Engine: realEngine}

	return fixture{
		service:      New(engine),
		engine:       engine,
		tasksStorage: tasksStorage,
		jobsStorage:  jobsStorage,
	}
}

func (f fixture) enqueue(t *testing.T, entityID string) tasks_queries.VelezTask {
	t.Helper()

	task, err := f.engine.Enqueue(context.Background(), entityID, testAction, &emptyContext{})
	require.NoError(t, err)

	return task
}

func (f fixture) addJob(t *testing.T, taskID int64, name string, status jobs_queries.VelezJobStatus) {
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

func (f fixture) fail(t *testing.T, taskID int64) {
	t.Helper()

	params := tasks_queries.FinishTaskParams{
		Status: tasks_queries.VelezTaskStatusFAILED,
		Error:  sql.NullString{String: failureText, Valid: true},
		ID:     taskID,
	}

	err := f.tasksStorage.FinishTask(context.Background(), params)
	require.NoError(t, err)
}

func Test_List_MapsTaskStatusErrorAndJobs(t *testing.T) {
	f := newFixture()
	task := f.enqueue(t, "e1")
	f.addJob(t, task.ID, firstJob, jobs_queries.VelezJobStatusDONE)
	f.addJob(t, task.ID, secondJob, jobs_queries.VelezJobStatusFAILED)
	f.fail(t, task.ID)

	got, err := f.service.List(context.Background(), testAction)
	require.NoError(t, err)

	require.Len(t, got, 1)
	require.Equal(t, task.ID, got[0].TaskId)
	require.Equal(t, "e1", got[0].EntityId)
	require.Equal(t, testAction, got[0].Action)
	require.Equal(t, velez_api.ProvisioningTask_FAILED, got[0].Status)
	require.Equal(t, failureText, got[0].Error)
	require.False(t, got[0].UpdatedAt.IsZero())

	jobStatuses := make(map[string]velez_api.ProvisioningTask_Status, len(got[0].Jobs))
	for _, job := range got[0].Jobs {
		jobStatuses[job.Name] = job.Status
	}

	require.Equal(t, map[string]velez_api.ProvisioningTask_Status{
		firstJob:  velez_api.ProvisioningTask_DONE,
		secondJob: velez_api.ProvisioningTask_FAILED,
		thirdJob:  velez_api.ProvisioningTask_PENDING,
	}, jobStatuses)
}

func Test_List_MapsRunningJobAndPendingTask(t *testing.T) {
	f := newFixture()
	task := f.enqueue(t, "e1")
	f.addJob(t, task.ID, firstJob, jobs_queries.VelezJobStatusRUNNING)

	got, err := f.service.List(context.Background(), testAction)
	require.NoError(t, err)

	require.Len(t, got, 1)
	require.Equal(t, velez_api.ProvisioningTask_PENDING, got[0].Status)
	require.Equal(t, velez_api.ProvisioningTask_RUNNING, got[0].Jobs[0].Status)
}

func Test_List_AppliesFiveMinuteFailedWindow(t *testing.T) {
	f := newFixture()

	before := time.Now()
	_, err := f.service.List(context.Background(), testAction)
	require.NoError(t, err)

	after := time.Now()

	require.False(t, f.engine.failedSince.Before(before.Add(-5*time.Minute)))
	require.False(t, f.engine.failedSince.After(after.Add(-5*time.Minute)))
}

func Test_Dismiss_DeletesFailedTask(t *testing.T) {
	f := newFixture()
	task := f.enqueue(t, "e1")
	f.fail(t, task.ID)

	err := f.service.Dismiss(context.Background(), "e1", testAction)
	require.NoError(t, err)

	got, err := f.service.List(context.Background(), testAction)
	require.NoError(t, err)
	require.Empty(t, got)
}

func Test_Dismiss_PropagatesNotDismissable(t *testing.T) {
	f := newFixture()
	f.enqueue(t, "e1")

	err := f.service.Dismiss(context.Background(), "e1", testAction)
	require.ErrorIs(t, err, user_errors.ErrTaskNotDismissable)
}
