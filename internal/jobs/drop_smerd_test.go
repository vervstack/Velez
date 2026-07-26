package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/sqlc-dev/pqtype"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

const (
	testUUID1 = "uuid-1"
	testUUID2 = "uuid-2"
)

// dockerOnlyNodeClients wraps an arbitrary node_clients.Docker (including
// wrapper fakes like selectiveFailDocker below, which fakeNodeClients can't
// hold since it's typed to a concrete *fakeDocker) into a full
// node_clients.NodeClients, satisfying dropSmerdHandler's only dependency.
type dockerOnlyNodeClients struct {
	node_clients.NodeClients
	docker node_clients.Docker
}

func (f *dockerOnlyNodeClients) Docker() node_clients.Docker {
	return f.docker
}

func dropSmerdTask(
	t *testing.T, tasksStorage *fakeTasksStorage, entityID string, req *velez_api.DropSmerd_Request,
) tasks_queries.VelezTask {
	t.Helper()

	initialContext := &velez_api.DropSmerdTaskPayload{}
	initialContext.SetRequest(req)

	payloadJSON, err := json.Marshal(initialContext)
	if err != nil {
		t.Fatalf("unexpected error marshaling payload: %v", err)
	}

	taskContext := pqtype.NullRawMessage{RawMessage: payloadJSON, Valid: true}
	createParams := tasks_queries.CreateTaskParams{
		EntityID: entityID,
		Action:   DropSmerdAction,
		Context:  taskContext,
	}

	task, err := tasksStorage.CreateTask(context.Background(), createParams)
	if err != nil {
		t.Fatalf("unexpected error creating task: %v", err)
	}

	return task
}

func TestDropSmerdHandler_Action(t *testing.T) {
	h := NewDropSmerdHandler(&dockerOnlyNodeClients{docker: newFakeDocker()})

	if h.Action() != DropSmerdAction {
		t.Errorf("expected action %q, got %q", DropSmerdAction, h.Action())
	}
}

func TestDropSmerdHandler_AllSucceed(t *testing.T) {
	tasksStorage := newFakeTasksStorage()
	jobsStorage := newFakeJobsStorage()
	docker := newFakeDocker()

	req := &velez_api.DropSmerd_Request{
		Uuids: []string{testUUID1, testUUID2},
		Name:  []string{"name-1"},
	}
	task := dropSmerdTask(t, tasksStorage, "batch-1", req)

	registry := NewRegistry()
	registry.Register(NewDropSmerdHandler(&dockerOnlyNodeClients{docker: docker}))

	w, ok := NewTaskWorker(tasksStorage, jobsStorage, registry, "test-worker", time.Hour).(*taskWorker)
	if !ok {
		t.Fatal("expected NewTaskWorker to return *taskWorker")
	}

	err := w.run(context.Background(), task)
	if err != nil {
		t.Fatalf("unexpected error running task: %v", err)
	}

	finished := tasksStorage.get(task.ID)
	if finished.Status != tasks_queries.VelezTaskStatusDONE {
		t.Fatalf("expected task status DONE, got %v", finished.Status)
	}

	var result velez_api.DropSmerdTaskPayload

	err = json.Unmarshal(finished.Context.RawMessage, &result)
	if err != nil {
		t.Fatalf("unexpected error unmarshaling task context: %v", err)
	}

	if len(result.GetFailed()) != 0 {
		t.Errorf("expected no failures, got %v", result.GetFailed())
	}

	wantSuccessful := []string{testUUID1, testUUID2, "name-1"}

	gotSuccessful := result.GetSuccessful()
	if len(gotSuccessful) != len(wantSuccessful) {
		t.Fatalf("expected successful %v, got %v", wantSuccessful, gotSuccessful)
	}

	for i, want := range wantSuccessful {
		if gotSuccessful[i] != want {
			t.Errorf("expected successful[%d] = %q, got %q", i, want, gotSuccessful[i])
		}
	}

	if len(docker.removeCalledWith) != 3 {
		t.Errorf("expected 3 Remove calls, got %v", docker.removeCalledWith)
	}
}

// TestDropSmerdHandler_PartialFailureStillReachesDone is the most important
// test in this file: it verifies the behavior-preservation guarantee that
// motivated dropContainerJob.Do's "always return nil" design - a single
// container's removal error must not fail the whole task, must not be
// surfaced as a top-level error, and must land only in the response's
// Failed slice, exactly like the old container_manager.DropSmerds RPC.
func TestDropSmerdHandler_PartialFailureStillReachesDone(t *testing.T) {
	tasksStorage := newFakeTasksStorage()
	jobsStorage := newFakeJobsStorage()
	docker := newFakeDocker()

	failingUUID := "uuid-bad"

	req := &velez_api.DropSmerd_Request{
		Uuids: []string{"uuid-good-1", failingUUID, "uuid-good-2"},
	}
	task := dropSmerdTask(t, tasksStorage, "batch-2", req)

	wrapped := &selectiveFailDocker{fakeDocker: docker, failOn: failingUUID}

	registry := NewRegistry()
	registry.Register(NewDropSmerdHandler(&dockerOnlyNodeClients{docker: wrapped}))

	w, ok := NewTaskWorker(tasksStorage, jobsStorage, registry, "test-worker", time.Hour).(*taskWorker)
	if !ok {
		t.Fatal("expected NewTaskWorker to return *taskWorker")
	}

	err := w.run(context.Background(), task)
	if err != nil {
		t.Fatalf("unexpected error running task: %v", err)
	}

	finished := tasksStorage.get(task.ID)
	if finished.Status != tasks_queries.VelezTaskStatusDONE {
		t.Fatalf("expected task status DONE despite a per-item failure, got %v", finished.Status)
	}

	var result velez_api.DropSmerdTaskPayload

	err = json.Unmarshal(finished.Context.RawMessage, &result)
	if err != nil {
		t.Fatalf("unexpected error unmarshaling task context: %v", err)
	}

	failed := result.GetFailed()
	if len(failed) != 1 {
		t.Fatalf("expected exactly one failure, got %v", failed)
	}

	if failed[0].GetUuid() != failingUUID {
		t.Errorf("expected failed uuid %q, got %q", failingUUID, failed[0].GetUuid())
	}

	if failed[0].GetCause() == "" {
		t.Errorf("expected a non-empty failure cause")
	}

	wantSuccessful := []string{"uuid-good-1", "uuid-good-2"}

	gotSuccessful := result.GetSuccessful()
	if len(gotSuccessful) != len(wantSuccessful) {
		t.Fatalf("expected successful %v, got %v", wantSuccessful, gotSuccessful)
	}

	for i, want := range wantSuccessful {
		if gotSuccessful[i] != want {
			t.Errorf("expected successful[%d] = %q, got %q", i, want, gotSuccessful[i])
		}
	}
}

func TestDropSmerdHandler_ResumeSkipsAlreadyDoneJobs(t *testing.T) {
	tasksStorage := newFakeTasksStorage()
	jobsStorage := newFakeJobsStorage()
	docker := newFakeDocker()

	req := &velez_api.DropSmerd_Request{
		Uuids: []string{testUUID1, testUUID2},
	}
	task := dropSmerdTask(t, tasksStorage, "batch-3", req)

	// Simulate a worker that already completed drop_container_0 (uuid-1) in
	// a previous, crashed attempt.
	jobsStorage.seedDone(task.ID, "drop_container_0")

	registry := NewRegistry()
	registry.Register(NewDropSmerdHandler(&dockerOnlyNodeClients{docker: docker}))

	w, ok := NewTaskWorker(tasksStorage, jobsStorage, registry, "test-worker", time.Hour).(*taskWorker)
	if !ok {
		t.Fatal("expected NewTaskWorker to return *taskWorker")
	}

	err := w.run(context.Background(), task)
	if err != nil {
		t.Fatalf("unexpected error running task: %v", err)
	}

	finished := tasksStorage.get(task.ID)
	if finished.Status != tasks_queries.VelezTaskStatusDONE {
		t.Fatalf("expected task status DONE, got %v", finished.Status)
	}

	// Only drop_container_1 (uuid-2) should have actually called Remove -
	// drop_container_0 was already DONE and must be skipped.
	if len(docker.removeCalledWith) != 1 || docker.removeCalledWith[0] != testUUID2 {
		t.Errorf("expected only uuid-2 to be removed, got %v", docker.removeCalledWith)
	}
}

// selectiveFailDocker wraps a *fakeDocker and makes Remove fail only for one
// specific identifier, succeeding (recording the call, via the embedded
// fake) for everything else - used to simulate one container failing
// mid-batch while others succeed.
type selectiveFailDocker struct {
	*fakeDocker
	failOn string
}

var errSelectiveFailDockerRemove = errors.New("error removing container: simulated failure")

func (f *selectiveFailDocker) Remove(ctx context.Context, uuid string) error {
	_ = f.fakeDocker.Remove(ctx, uuid)
	if uuid == f.failOn {
		return errSelectiveFailDockerRemove
	}

	return nil
}
