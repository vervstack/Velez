package local_storage

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

func newTaskParams() tasks_queries.CreateTaskParams {
	return tasks_queries.CreateTaskParams{EntityID: "entity", Action: "action"}
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

			lookupReq := tasks_queries.GetTaskByEntityActionParams{EntityID: "entity", Action: "action"}

			newest, err := storage.GetTaskByEntityAction(ctx, lookupReq)
			require.NoError(t, err)
			require.Equal(t, second.ID, newest.ID)
		})
	}
}

func Test_Tasks_SubscribeReceivesSignalAfterNotify(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
