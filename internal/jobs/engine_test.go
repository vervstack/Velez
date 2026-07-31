package jobs

import (
	"context"
	"sync"
	"testing"

	"go.vervstack.ru/Velez/internal/storage/postgres/generated/jobs_queries"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

func TestEngine_EnqueueDedupesConcurrentCalls(t *testing.T) {
	tasksStorage := newFakeTasksStorage()
	jobsStorage := newFakeJobsStorage()
	engine := NewEngine(tasksStorage, jobsStorage)

	const n = 20

	ids := make([]int64, n)
	errs := make([]error, n)

	var wg sync.WaitGroup

	wg.Add(n)

	for i := range n {
		go func(i int) {
			defer wg.Done()

			task, err := engine.Enqueue(context.Background(), "same-entity", "same-action", &dummyContext{Value: "x"})

			ids[i] = task.ID
			errs[i] = err
		}(i)
	}

	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("call %d: unexpected error: %v", i, err)
		}
	}

	first := ids[0]
	for i, id := range ids {
		if id != first {
			t.Errorf("call %d returned task id %d, expected all calls to converge on %d", i, id, first)
		}
	}

	tasksStorage.mu.Lock()

	count := len(tasksStorage.byID)
	tasksStorage.mu.Unlock()

	if count != 1 {
		t.Errorf("expected exactly 1 task to be created, got %d", count)
	}
}

func TestEngine_EnqueueReturnsExistingFailedTaskInstead(t *testing.T) {
	tasksStorage := newFakeTasksStorage()
	jobsStorage := newFakeJobsStorage()
	engine := NewEngine(tasksStorage, jobsStorage)

	first, err := engine.Enqueue(context.Background(), "e", "a", &dummyContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tasksStorage.mu.Lock()

	failed := tasksStorage.byID[first.ID]

	failed.Status = tasks_queries.VelezTaskStatusFAILED
	tasksStorage.byID[first.ID] = failed
	tasksStorage.mu.Unlock()

	second, err := engine.Enqueue(context.Background(), "e", "a", &dummyContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if second.ID != first.ID {
		t.Errorf("expected Enqueue to return the existing task, got a different id")
	}

	if second.Status != tasks_queries.VelezTaskStatusFAILED {
		t.Errorf("expected caller to see the existing FAILED status, got %v", second.Status)
	}
}

const (
	// listJobsTestAction is the action name testHandler (see worker_test.go)
	// registers ListJobs's fixture tasks under.
	listJobsTestAction = "list_jobs_test_action"
	listJobsTestEntity = "list_jobs_test_entity"

	jobNameA = "job_a"
	jobNameB = "job_b"
	jobNameC = "job_c"
	jobNameD = "job_d"
)

// newListJobsFixture builds an engine wired to a registry with a single
// testHandler whose BuildJobs returns four jobs in a fixed order - used to
// exercise ListJobs's ordering/merge behavior below.
func newListJobsFixture() (*engine, *fakeTasksStorage, *fakeJobsStorage) {
	tasksStorage := newFakeTasksStorage()
	jobsStorage := newFakeJobsStorage()

	e, ok := NewEngine(tasksStorage, jobsStorage).(*engine)

	if !ok {
		panic("NewEngine did not return *engine")
	}

	registry := NewRegistry()
	registry.Register(&testHandler{
		action: listJobsTestAction,
		jobs: func(TaskContext) []NamedJob {
			return []NamedJob{
				{Name: jobNameA, Job: nil},
				{Name: jobNameB, Job: nil},
				{Name: jobNameC, Job: nil},
				{Name: jobNameD, Job: nil},
			}
		},
	})
	e.SetRegistry(registry)

	return e, tasksStorage, jobsStorage
}

func TestEngine_ListJobs_OrderedAndMergedWithPersistedStatus(t *testing.T) {
	e, tasksStorage, jobsStorage := newListJobsFixture()

	task, err := tasksStorage.CreateTask(context.Background(), tasks_queries.CreateTaskParams{
		EntityID: listJobsTestEntity,
		Action:   listJobsTestAction,
	})
	if err != nil {
		t.Fatalf("unexpected error creating task: %v", err)
	}

	// jobNameA: DONE, jobNameB: RUNNING, jobNameC: FAILED, jobNameD: no row (pending).
	_, err = jobsStorage.CreateRunningJob(context.Background(), jobs_queries.CreateRunningJobParams{
		TaskID: task.ID, JobName: jobNameA,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = jobsStorage.FinishJob(context.Background(), jobs_queries.FinishJobParams{
		TaskID: task.ID, JobName: jobNameA, Status: jobs_queries.VelezJobStatusDONE,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = jobsStorage.CreateRunningJob(context.Background(), jobs_queries.CreateRunningJobParams{
		TaskID: task.ID, JobName: jobNameB,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = jobsStorage.CreateRunningJob(context.Background(), jobs_queries.CreateRunningJobParams{
		TaskID: task.ID, JobName: jobNameC,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = jobsStorage.FinishJob(context.Background(), jobs_queries.FinishJobParams{
		TaskID: task.ID, JobName: jobNameC, Status: jobs_queries.VelezJobStatusFAILED,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	statuses, err := e.ListJobs(context.Background(), task)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []JobStatus{
		{Name: jobNameA, Status: jobs_queries.VelezJobStatusDONE},
		{Name: jobNameB, Status: jobs_queries.VelezJobStatusRUNNING},
		{Name: jobNameC, Status: jobs_queries.VelezJobStatusFAILED},
		{Name: jobNameD, Status: ""},
	}

	if len(statuses) != len(want) {
		t.Fatalf("expected %d statuses, got %d: %+v", len(want), len(statuses), statuses)
	}

	for i, w := range want {
		if statuses[i] != w {
			t.Errorf("index %d: expected %+v, got %+v", i, w, statuses[i])
		}
	}
}

func TestEngine_ListJobs_NoHandlerRegistered(t *testing.T) {
	tasksStorage := newFakeTasksStorage()
	jobsStorage := newFakeJobsStorage()
	e, ok := NewEngine(tasksStorage, jobsStorage).(*engine)

	if !ok {
		t.Fatal("NewEngine did not return *engine")
	}

	e.SetRegistry(NewRegistry())

	task, err := tasksStorage.CreateTask(context.Background(), tasks_queries.CreateTaskParams{
		EntityID: listJobsTestEntity,
		Action:   "unregistered_action",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	statuses, err := e.ListJobs(context.Background(), task)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if statuses != nil {
		t.Errorf("expected nil statuses when no handler is registered, got %+v", statuses)
	}
}
