package jobs

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

const (
	testJobNameFirst  = "first"
	testJobNameSecond = "second"
	testAction        = "test_action"
)

type orderedJob struct {
	name    string
	log     *[]string
	failure error
}

func (o *orderedJob) Do(_ context.Context) error {
	*o.log = append(*o.log, o.name)

	return o.failure
}

type testHandler struct {
	action string
	jobs   func(taskCtx TaskContext) []NamedJob
}

func (h *testHandler) Action() string                      { return h.action }
func (h *testHandler) NewContext() TaskContext             { return &dummyContext{} }
func (h *testHandler) BuildJobs(tc TaskContext) []NamedJob { return h.jobs(tc) }

func TestTaskWorker_ClaimsAndRunsAllJobsInOrder(t *testing.T) {
	tasksStorage := newFakeTasksStorage()
	jobsStorage := newFakeJobsStorage()

	createParams := tasks_queries.CreateTaskParams{EntityID: "e1", Action: testAction}

	task, err := tasksStorage.CreateTask(context.Background(), createParams)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var log []string

	registry := NewRegistry()
	registry.Register(&testHandler{
		action: testAction,
		jobs: func(TaskContext) []NamedJob {
			return []NamedJob{
				{Name: testJobNameFirst, Job: &orderedJob{name: testJobNameFirst, log: &log}},
				{Name: testJobNameSecond, Job: &orderedJob{name: testJobNameSecond, log: &log}},
			}
		},
	})

	w, ok := NewTaskWorker(tasksStorage, jobsStorage, registry, "test-worker", time.Hour).(*taskWorker)
	if !ok {
		t.Fatal("expected NewTaskWorker to return *taskWorker")
	}

	w.processOne(context.Background())

	if len(log) != 2 || log[0] != testJobNameFirst || log[1] != testJobNameSecond {
		t.Errorf("expected jobs to run in order [first second], got %v", log)
	}

	finished := tasksStorage.get(task.ID)
	if finished.Status != tasks_queries.VelezTaskStatusDONE {
		t.Errorf("expected task status DONE, got %v", finished.Status)
	}
}

func TestTaskWorker_ReclaimsStaleRunningTaskAndSkipsDoneJobs(t *testing.T) {
	tasksStorage := newFakeTasksStorage()
	jobsStorage := newFakeJobsStorage()

	createParams := tasks_queries.CreateTaskParams{EntityID: "e2", Action: testAction}

	task, err := tasksStorage.CreateTask(context.Background(), createParams)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Simulate a worker that claimed this task, completed "first", then crashed
	// before testJobNameSecond ran: claimed_at is stale and "first" is already DONE.
	tasksStorage.mu.Lock()
	stale := tasksStorage.byID[task.ID]
	stale.Status = tasks_queries.VelezTaskStatusRUNNING
	stale.ClaimedAt = sql.NullTime{Time: time.Now().Add(-time.Hour), Valid: true}
	stale.ClaimedBy = sql.NullString{String: "dead-worker", Valid: true}
	tasksStorage.byID[task.ID] = stale
	tasksStorage.mu.Unlock()

	jobsStorage.seedDone(task.ID, testJobNameFirst)

	var log []string

	registry := NewRegistry()
	registry.Register(&testHandler{
		action: testAction,
		jobs: func(TaskContext) []NamedJob {
			return []NamedJob{
				{Name: testJobNameFirst, Job: &orderedJob{name: testJobNameFirst, log: &log}},
				{Name: testJobNameSecond, Job: &orderedJob{name: testJobNameSecond, log: &log}},
			}
		},
	})

	w, ok := NewTaskWorker(tasksStorage, jobsStorage, registry, "new-worker", time.Hour).(*taskWorker)
	if !ok {
		t.Fatal("expected NewTaskWorker to return *taskWorker")
	}

	w.processOne(context.Background())

	if len(log) != 1 || log[0] != testJobNameSecond {
		t.Errorf("expected only the not-yet-done job to run, got %v", log)
	}

	finished := tasksStorage.get(task.ID)
	if finished.Status != tasks_queries.VelezTaskStatusDONE {
		t.Errorf("expected task status DONE after reclaim, got %v", finished.Status)
	}
}
