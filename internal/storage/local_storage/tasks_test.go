package local_storage

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

const (
	testTaskAction = "action"
)

func newTaskParams() tasks_queries.CreateTaskParams {
	return tasks_queries.CreateTaskParams{EntityID: "entity", Action: testTaskAction}
}

func Test_Tasks_CreateWhileInFlightReturnsNoRows(t *testing.T) {
	storage := newTasksStorage()

	_, err := storage.CreateTask(context.Background(), newTaskParams())
	require.NoError(t, err)

	_, err = storage.CreateTask(context.Background(), newTaskParams())
	require.ErrorIs(t, err, sql.ErrNoRows)
}

func Test_Tasks_CreateAfterFinishedTaskGetsNewTask(t *testing.T) {
	cases := []struct {
		name           string
		finishedStatus tasks_queries.VelezTaskStatus
	}{
		{"done task", tasks_queries.VelezTaskStatusDONE},
		{"failed task", tasks_queries.VelezTaskStatusFAILED},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			storage := newTasksStorage()

			first, err := storage.CreateTask(ctx, newTaskParams())
			require.NoError(t, err)

			finishReq := tasks_queries.FinishTaskParams{ID: first.ID, Status: tc.finishedStatus}
			require.NoError(t, storage.FinishTask(ctx, finishReq))

			second, err := storage.CreateTask(ctx, newTaskParams())
			require.NoError(t, err)
			require.NotEqual(t, first.ID, second.ID)

			lookupReq := tasks_queries.GetTaskByEntityActionParams{EntityID: "entity", Action: testTaskAction}

			newest, err := storage.GetTaskByEntityAction(ctx, lookupReq)
			require.NoError(t, err)
			require.Equal(t, second.ID, newest.ID)
		})
	}
}

func Test_Tasks_SubscribeReceivesSignalAfterNotify(t *testing.T) {
	ctx := t.Context()

	storage := newTasksStorage()
	wake := storage.SubscribeTaskCreated(ctx)

	err := storage.NotifyTaskCreated(ctx)
	require.NoError(t, err)

	select {
	case <-wake:
	case <-time.After(time.Second):
		t.Fatal("no wake signal after notify")
	}
}

func Test_Tasks_SubscribeStopsAfterContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	storage := newTasksStorage()
	wake := storage.SubscribeTaskCreated(ctx)

	cancel()

	require.Eventually(t, func() bool {
		select {
		case _, isOpen := <-wake:
			return !isOpen
		default:
			return false
		}
	}, time.Second, 10*time.Millisecond)

	err := storage.NotifyTaskCreated(context.Background())
	require.NoError(t, err)
}

func newTaskWithStatus(t *testing.T, storage *tasks, entityID string, status tasks_queries.VelezTaskStatus) int64 {
	t.Helper()

	ctx := context.Background()

	task, err := storage.CreateTask(ctx, tasks_queries.CreateTaskParams{EntityID: entityID, Action: testTaskAction})
	require.NoError(t, err)

	if status == tasks_queries.VelezTaskStatusPENDING {
		return task.ID
	}

	_, err = storage.ClaimTask(ctx, tasks_queries.ClaimTaskParams{
		ClaimedAt: sql.NullTime{Time: time.Now().Add(-time.Hour), Valid: true},
	})
	require.NoError(t, err)

	if status == tasks_queries.VelezTaskStatusRUNNING {
		return task.ID
	}

	err = storage.FinishTask(ctx, tasks_queries.FinishTaskParams{ID: task.ID, Status: status})
	require.NoError(t, err)

	return task.ID
}

func Test_Tasks_EvictFinishedDropsOnlyOldDoneTasks(t *testing.T) {
	ctx := context.Background()
	storage := newTasksStorage()

	doneId := newTaskWithStatus(t, storage, "done", tasks_queries.VelezTaskStatusDONE)
	failedId := newTaskWithStatus(t, storage, "failed", tasks_queries.VelezTaskStatusFAILED)
	runningId := newTaskWithStatus(t, storage, "running", tasks_queries.VelezTaskStatusRUNNING)
	pendingId := newTaskWithStatus(t, storage, "pending", tasks_queries.VelezTaskStatusPENDING)

	evicted, err := storage.EvictFinished(ctx, time.Now().Add(time.Hour))
	require.NoError(t, err)
	require.Equal(t, []int64{doneId}, evicted)

	_, err = storage.GetTaskById(ctx, doneId)
	require.ErrorIs(t, err, sql.ErrNoRows)

	for _, keptId := range []int64{failedId, runningId, pendingId} {
		_, err = storage.GetTaskById(ctx, keptId)
		require.NoError(t, err)
	}
}

func Test_Tasks_EvictFinishedKeepsDoneTaskNewerThanCutoff(t *testing.T) {
	ctx := context.Background()
	storage := newTasksStorage()

	doneId := newTaskWithStatus(t, storage, "done", tasks_queries.VelezTaskStatusDONE)

	evicted, err := storage.EvictFinished(ctx, time.Now().Add(-time.Hour))
	require.NoError(t, err)
	require.Empty(t, evicted)

	_, err = storage.GetTaskById(ctx, doneId)
	require.NoError(t, err)
}
