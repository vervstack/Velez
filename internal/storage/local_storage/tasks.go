package local_storage

import (
	"context"
	"database/sql"
	"slices"
	"sort"
	"sync"
	"time"

	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

// tasks is a real, in-memory implementation of storage.TasksStorage, used
// when Velez runs without Postgres configured. It mirrors the Postgres
// queries' semantics (dedup against in-flight tasks on entity_id+action,
// newest-task lookup, atomic claim, stale reclaim) closely enough for the task/job engine (internal/jobs) to work
// correctly - state just doesn't survive a process restart in this mode.
type tasks struct {
	mu     sync.Mutex
	nextID int64
	byID   map[int64]tasks_queries.VelezTask
}

func NewTasksStorage() storage.TasksStorage {
	return newTasksStorage()
}

func newTasksStorage() *tasks {
	return &tasks{
		byID: make(map[int64]tasks_queries.VelezTask),
	}
}

func (t *tasks) CreateTask(_ context.Context, arg tasks_queries.CreateTaskParams) (tasks_queries.VelezTask, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, existing := range t.byID {
		isInFlight := existing.Status == tasks_queries.VelezTaskStatusPENDING ||
			existing.Status == tasks_queries.VelezTaskStatusRUNNING

		if existing.EntityID == arg.EntityID && existing.Action == arg.Action && isInFlight {
			return tasks_queries.VelezTask{}, sql.ErrNoRows
		}
	}

	t.nextID++

	now := time.Now()

	task := tasks_queries.VelezTask{
		ID:        t.nextID,
		EntityID:  arg.EntityID,
		Action:    arg.Action,
		Status:    tasks_queries.VelezTaskStatusPENDING,
		Context:   arg.Context,
		CreatedAt: now,
		UpdatedAt: now,
	}

	t.byID[task.ID] = task

	return task, nil
}

func (t *tasks) GetTaskByEntityAction(_ context.Context,
	arg tasks_queries.GetTaskByEntityActionParams,
) (tasks_queries.VelezTask, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	var newest tasks_queries.VelezTask

	isFound := false

	for _, task := range t.byID {
		if task.EntityID == arg.EntityID && task.Action == arg.Action && (!isFound || task.ID > newest.ID) {
			newest = task
			isFound = true
		}
	}

	if !isFound {
		return tasks_queries.VelezTask{}, sql.ErrNoRows
	}

	return newest, nil
}

func (t *tasks) GetTaskById(_ context.Context, id int64) (tasks_queries.VelezTask, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	task, ok := t.byID[id]
	if !ok {
		return tasks_queries.VelezTask{}, sql.ErrNoRows
	}

	return task, nil
}

func (t *tasks) ClaimTask(_ context.Context, arg tasks_queries.ClaimTaskParams) (tasks_queries.VelezTask, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	staleThreshold := arg.ClaimedAt.Time

	for id, task := range t.byID {
		claimable := task.Status == tasks_queries.VelezTaskStatusPENDING ||
			(task.Status == tasks_queries.VelezTaskStatusRUNNING && task.ClaimedAt.Valid &&
				task.ClaimedAt.Time.Before(staleThreshold))
		if !claimable {
			continue
		}

		task.Status = tasks_queries.VelezTaskStatusRUNNING
		task.ClaimedAt = sql.NullTime{Time: time.Now(), Valid: true}
		task.ClaimedBy = arg.ClaimedBy
		task.UpdatedAt = time.Now()
		t.byID[id] = task

		return task, nil
	}

	return tasks_queries.VelezTask{}, sql.ErrNoRows
}

func (t *tasks) UpdateTaskContext(_ context.Context, arg tasks_queries.UpdateTaskContextParams) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	task, ok := t.byID[arg.ID]
	if !ok {
		return sql.ErrNoRows
	}

	task.Context = arg.Context
	task.UpdatedAt = time.Now()
	t.byID[arg.ID] = task

	return nil
}

func (t *tasks) RenewTaskClaim(_ context.Context, arg tasks_queries.RenewTaskClaimParams) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	task, ok := t.byID[arg.ID]
	if !ok || task.Status != tasks_queries.VelezTaskStatusRUNNING || task.ClaimedBy != arg.ClaimedBy {
		return nil
	}

	task.ClaimedAt = sql.NullTime{Time: time.Now(), Valid: true}
	t.byID[arg.ID] = task

	return nil
}

func (t *tasks) FinishTask(_ context.Context, arg tasks_queries.FinishTaskParams) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	task, ok := t.byID[arg.ID]
	if !ok {
		return sql.ErrNoRows
	}

	task.Status = arg.Status
	task.Error = arg.Error
	task.UpdatedAt = time.Now()
	t.byID[arg.ID] = task

	return nil
}

func (t *tasks) ListProvisioningTasks(_ context.Context,
	arg tasks_queries.ListProvisioningTasksParams,
) ([]tasks_queries.VelezTask, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	out := make([]tasks_queries.VelezTask, 0)

	for _, task := range t.byID {
		if !slices.Contains(arg.Actions, task.Action) {
			continue
		}

		isInFlight := task.Status == tasks_queries.VelezTaskStatusPENDING ||
			task.Status == tasks_queries.VelezTaskStatusRUNNING
		isRecentlyFailed := task.Status == tasks_queries.VelezTaskStatusFAILED &&
			task.UpdatedAt.After(arg.FailedSince)

		if isInFlight || isRecentlyFailed {
			out = append(out, task)
		}
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})

	return out, nil
}

func (t *tasks) DeleteTask(_ context.Context, id int64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.byID, id)

	return nil
}

func (t *tasks) WithTx(_ *sql.Tx) *tasks_queries.Queries {
	return nil
}
