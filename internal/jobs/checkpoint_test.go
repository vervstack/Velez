package jobs

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"

	"go.vervstack.ru/Velez/internal/storage/postgres/generated/jobs_queries"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

type dummyContext struct {
	Value string
}

type recordingJob struct {
	called bool
	doFunc func(ctx context.Context) error
}

func (r *recordingJob) Do(ctx context.Context) error {
	r.called = true
	if r.doFunc != nil {
		return r.doFunc(ctx)
	}

	return nil
}

func TestCheckpoint_SkipsWhenDone(t *testing.T) {
	jobsStorage := newFakeJobsStorage()
	tasksStorage := newFakeTasksStorage()

	jobsStorage.seedDone(1, "my_job")

	inner := &recordingJob{}
	taskCtx := &dummyContext{Value: "unchanged"}

	err := Checkpoint(jobsStorage, tasksStorage, 1, "my_job", taskCtx, inner).Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if inner.called {
		t.Error("expected inner job not to run when checkpoint is already DONE")
	}
}

func TestCheckpoint_FailFastWhenFailed(t *testing.T) {
	jobsStorage := newFakeJobsStorage()
	tasksStorage := newFakeTasksStorage()

	jobsStorage.rows[jobKey(1, "my_job")] = jobs_queries.VelezJob{
		TaskID:  1,
		JobName: "my_job",
		Status:  jobs_queries.VelezJobStatusFAILED,
		Error:   sql.NullString{String: "boom", Valid: true},
	}

	inner := &recordingJob{}
	taskCtx := &dummyContext{}

	err := Checkpoint(jobsStorage, tasksStorage, 1, "my_job", taskCtx, inner).Do(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "boom") {
		t.Errorf("expected stored failure reason in error, got: %v", err)
	}

	if inner.called {
		t.Error("expected inner job not to run when checkpoint already FAILED")
	}
}

func TestCheckpoint_RunsAndPersistsContextOnSuccess(t *testing.T) {
	jobsStorage := newFakeJobsStorage()
	tasksStorage := newFakeTasksStorage()

	createParams := tasks_queries.CreateTaskParams{EntityID: "entity", Action: "action"}

	task, err := tasksStorage.CreateTask(context.Background(), createParams)
	if err != nil {
		t.Fatalf("unexpected error seeding task: %v", err)
	}

	taskCtx := &dummyContext{Value: "initial"}

	inner := &recordingJob{
		doFunc: func(_ context.Context) error {
			taskCtx.Value = "mutated"

			return nil
		},
	}

	err = Checkpoint(jobsStorage, tasksStorage, task.ID, "my_job", taskCtx, inner).Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !inner.called {
		t.Error("expected inner job to run")
	}

	row, err := jobsStorage.GetJob(context.Background(), jobs_queries.GetJobParams{TaskID: task.ID, JobName: "my_job"})
	if err != nil {
		t.Fatalf("expected job checkpoint row to exist: %v", err)
	}

	if row.Status != jobs_queries.VelezJobStatusDONE {
		t.Errorf("expected job status DONE, got %v", row.Status)
	}

	persisted := tasksStorage.get(task.ID)
	if !persisted.Context.Valid {
		t.Fatal("expected task context to be persisted")
	}

	var got dummyContext

	err = json.Unmarshal(persisted.Context.RawMessage, &got)
	if err != nil {
		t.Fatalf("error unmarshaling persisted context: %v", err)
	}

	if got.Value != "mutated" {
		t.Errorf("expected persisted context to reflect job's mutation, got %q", got.Value)
	}
}
