package jobs

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/sqlc-dev/pqtype"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

const (
	testCreateServiceName = "my_service"
)

// stubServicesResolver re-serves the same services storage per call, standing
// in for the swappable cluster state manager the create_service handler now
// holds.
type stubServicesResolver struct{ s storage.ServicesStorage }

func (r stubServicesResolver) Services() storage.ServicesStorage { return r.s }

func createServiceTask(t *testing.T, tasksStorage *fakeTasksStorage, entityID, name string) tasks_queries.VelezTask {
	t.Helper()

	payloadJSON, err := json.Marshal(&velez_api.CreateServiceTaskPayload{Name: name})
	if err != nil {
		t.Fatalf("unexpected error marshaling payload: %v", err)
	}

	task, err := tasksStorage.CreateTask(context.Background(), tasks_queries.CreateTaskParams{
		EntityID: entityID,
		Action:   CreateServiceAction,
		Context:  pqtype.NullRawMessage{RawMessage: payloadJSON, Valid: true},
	})
	if err != nil {
		t.Fatalf("unexpected error creating task: %v", err)
	}

	return task
}

func TestCreateServiceHandler_Action(t *testing.T) {
	h := NewCreateServiceHandler(stubServicesResolver{s: newFakeServicesStorage()})

	if h.Action() != CreateServiceAction {
		t.Errorf("expected action %q, got %q", CreateServiceAction, h.Action())
	}
}

func TestCreateServiceHandler_ValidNameUpsertsService(t *testing.T) {
	tasksStorage := newFakeTasksStorage()
	jobsStorage := newFakeJobsStorage()
	servicesStorage := newFakeServicesStorage()

	task := createServiceTask(t, tasksStorage, testCreateServiceName, testCreateServiceName)

	registry := NewRegistry()
	registry.Register(NewCreateServiceHandler(stubServicesResolver{s: servicesStorage}))

	w := NewTaskWorker(tasksStorage, jobsStorage, registry, "test-worker", time.Hour)

	err := w.run(context.Background(), task)
	if err != nil {
		t.Fatalf("unexpected error running task: %v", err)
	}

	if len(servicesStorage.upserted) != 1 || servicesStorage.upserted[0] != testCreateServiceName {
		t.Errorf("expected service 'my_service' to be upserted, got %v", servicesStorage.upserted)
	}

	finished := tasksStorage.get(task.ID)
	if finished.Status != tasks_queries.VelezTaskStatusDONE {
		t.Errorf("expected task status DONE, got %v", finished.Status)
	}
}

func TestCreateServiceHandler_InvalidNameFailsWithoutUpsert(t *testing.T) {
	tasksStorage := newFakeTasksStorage()
	jobsStorage := newFakeJobsStorage()
	servicesStorage := newFakeServicesStorage()

	// "bad" is shorter than the minimum allowed service name length.
	task := createServiceTask(t, tasksStorage, "bad", "bad")

	registry := NewRegistry()
	registry.Register(NewCreateServiceHandler(stubServicesResolver{s: servicesStorage}))

	w := NewTaskWorker(tasksStorage, jobsStorage, registry, "test-worker", time.Hour)

	err := w.run(context.Background(), task)
	if err == nil {
		t.Fatal("expected an error for an invalid service name")
	}

	if len(servicesStorage.upserted) != 0 {
		t.Errorf("expected no service to be upserted, got %v", servicesStorage.upserted)
	}

	finished := tasksStorage.get(task.ID)
	if finished.Status != tasks_queries.VelezTaskStatusFAILED {
		t.Errorf("expected task status FAILED, got %v", finished.Status)
	}
}
