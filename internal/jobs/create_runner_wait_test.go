package jobs

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	testWaitRunnerName = "gitlab_runner_wait"
)

func newWaitForRunnerDeployJob(watcher taskWatcher) *waitForRunnerDeployJob {
	payload := &velez_api.CreateRunnerTaskPayload{
		Request: &velez_api.CreateRunner_Request{Name: testWaitRunnerName},
	}

	return &waitForRunnerDeployJob{
		jobsEngine:   watcher,
		req:          payload,
		ctx:          payload,
		instanceName: testWaitRunnerName,
	}
}

func Test_WaitForRunnerDeployJob_StaleDoneTaskIgnoredFailedNewerTaskFails(t *testing.T) {
	watcher := &fakeTaskWatcher{
		tasks: []tasks_queries.VelezTask{
			{ID: 3, Status: tasks_queries.VelezTaskStatusDONE},
			{ID: 8, Status: tasks_queries.VelezTaskStatusFAILED, Error: sql.NullString{String: testErrorBoom, Valid: true}},
		},
	}
	job := newWaitForRunnerDeployJob(watcher)
	job.ctx.SetDeployBaselineTaskId(3)

	err := job.Do(context.Background())
	require.ErrorIs(t, err, user_errors.ErrTaskFailed)
}

func Test_WaitForRunnerDeployJob_ReturnsOnlyOnceNewerTaskIsDone(t *testing.T) {
	watcher := &fakeTaskWatcher{
		tasks: []tasks_queries.VelezTask{
			{ID: 3, Status: tasks_queries.VelezTaskStatusFAILED, Error: sql.NullString{String: testStaleOld, Valid: true}},
			{ID: 9, Status: tasks_queries.VelezTaskStatusRUNNING},
			{ID: 9, Status: tasks_queries.VelezTaskStatusDONE},
		},
	}
	job := newWaitForRunnerDeployJob(watcher)
	job.ctx.SetDeployBaselineTaskId(3)

	err := job.Do(context.Background())
	require.NoError(t, err)
}

func Test_RecordDeployBaseline_RunnerPayloadRecordedOnceAndNotOverwritten(t *testing.T) {
	watcher := &fakeTaskWatcher{latest: sql.Null[tasks_queries.VelezTask]{
		V:     tasks_queries.VelezTask{ID: 21},
		Valid: true,
	}}
	payload := &velez_api.CreateRunnerTaskPayload{}

	err := recordDeployBaseline(context.Background(), watcher, payload, "entity")
	require.NoError(t, err)

	watcher.latest.V.ID = 30

	err = recordDeployBaseline(context.Background(), watcher, payload, "entity")
	require.NoError(t, err)
	require.Equal(t, int64(21), payload.GetDeployBaselineTaskId())
	require.Equal(t, 1, watcher.latestCalls)
}
