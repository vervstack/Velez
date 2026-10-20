package jobs

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/service/user_errors"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/local_storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

type blockingJob struct {
	started chan struct{}
	release chan struct{}
}

func (b *blockingJob) Do(_ context.Context) error {
	close(b.started)
	<-b.release

	return nil
}

func newBlockingJob() *blockingJob {
	return &blockingJob{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
}

func newBlockingRegistry(job *blockingJob) *Registry {
	registry := NewRegistry()
	registry.Register(&testHandler{
		action: testAction,
		jobs: func(TaskContext) []NamedJob {
			return []NamedJob{{Name: testJobNameFirst, Job: job}}
		},
	})

	return registry
}

func startWorkerOnBlockedTask(t *testing.T, job *blockingJob) (*taskWorker, storage.TasksStorage) {
	t.Helper()

	tasksStorage := local_storage.NewTasksStorage()
	jobsStorage := local_storage.NewJobsStorage()

	w := NewTaskWorker(tasksStorage, jobsStorage, newBlockingRegistry(job), "drain-worker", time.Hour, 1)

	go w.Start(t.Context())

	require.Eventually(t, func() bool {
		return w.activeCount() == 0
	}, time.Second, 5*time.Millisecond)

	engine := NewEngine(tasksStorage, jobsStorage)

	_, err := engine.Enqueue(t.Context(), "e-drain", testAction, &dummyContext{})
	require.NoError(t, err)

	select {
	case <-job.started:
	case <-time.After(5 * time.Second):
		t.Fatal("job did not start")
	}

	return w, tasksStorage
}

func Test_Engine_BeginDrainRejectsExternalEnqueue(t *testing.T) {
	engine := NewEngine(local_storage.NewTasksStorage(), local_storage.NewJobsStorage())
	engine.BeginDrain()

	_, err := engine.Enqueue(context.Background(), "e1", testAction, &dummyContext{})
	require.ErrorIs(t, err, user_errors.ErrShuttingDown)

	_, err = engine.EnqueueReplacing(context.Background(), "e1", testAction, &dummyContext{})
	require.ErrorIs(t, err, user_errors.ErrShuttingDown)
}

func Test_Engine_BeginDrainAcceptsInternalEnqueue(t *testing.T) {
	engine := NewEngine(local_storage.NewTasksStorage(), local_storage.NewJobsStorage())
	engine.BeginDrain()

	ctx := WithInternalEnqueue(context.Background())

	_, err := engine.Enqueue(ctx, "e1", testAction, &dummyContext{})
	require.NoError(t, err)

	_, err = engine.EnqueueReplacing(ctx, "e2", testAction, &dummyContext{})
	require.NoError(t, err)
}

func Test_TaskWorker_StopWaitsForInFlightTask(t *testing.T) {
	job := newBlockingJob()
	w, tasksStorage := startWorkerOnBlockedTask(t, job)

	stopped := make(chan struct{})

	w.SetDrainTimeout(10 * time.Second)

	go func() {
		defer close(stopped)

		_ = w.Stop()
	}()

	select {
	case <-stopped:
		t.Fatal("Stop returned while a task was still running")
	case <-time.After(300 * time.Millisecond):
	}

	close(job.release)

	select {
	case <-stopped:
	case <-time.After(5 * time.Second):
		t.Fatal("Stop did not return after the task finished")
	}

	task, err := tasksStorage.GetTaskById(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, tasks_queries.VelezTaskStatusDONE, task.Status)
}

func Test_TaskWorker_StopGivesUpAfterDrainTimeout(t *testing.T) {
	job := newBlockingJob()
	w, _ := startWorkerOnBlockedTask(t, job)

	defer close(job.release)

	w.SetDrainTimeout(200 * time.Millisecond)

	startedAt := time.Now()

	err := w.Stop()
	require.NoError(t, err)
	require.Less(t, time.Since(startedAt), 3*time.Second)

	secondStartedAt := time.Now()

	err = w.Stop()
	require.NoError(t, err)
	require.Less(t, time.Since(secondStartedAt), 100*time.Millisecond)
}

func Test_TaskWorker_EvictsFinishedTasksOnFallbackTick(t *testing.T) {
	tasksStorage := local_storage.NewTasksStorage()
	jobsStorage := local_storage.NewJobsStorage()

	job := newBlockingJob()
	close(job.release)

	w := NewTaskWorker(tasksStorage, jobsStorage, newBlockingRegistry(job), "evict-worker", 20*time.Millisecond, 1)
	w.SetFinishedTaskTtl(time.Millisecond)

	defer func() { _ = w.Stop() }()

	go w.Start(t.Context())

	engine := NewEngine(tasksStorage, jobsStorage)

	task, err := engine.Enqueue(t.Context(), "e-evict", testAction, &dummyContext{})
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		_, getErr := tasksStorage.GetTaskById(context.Background(), task.ID)

		return errors.Is(getErr, sql.ErrNoRows)
	}, 5*time.Second, 10*time.Millisecond)

	jobs, err := jobsStorage.ListJobsByTask(context.Background(), task.ID)
	require.NoError(t, err)
	require.Empty(t, jobs)
}
