package jobs

import (
	"context"
	"sync"
	"testing"

	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

func TestEngine_EnqueueDedupesConcurrentCalls(t *testing.T) {
	tasksStorage := newFakeTasksStorage()
	engine := NewEngine(tasksStorage)

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
	engine := NewEngine(tasksStorage)

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
