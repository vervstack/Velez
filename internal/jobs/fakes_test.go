package jobs

import (
	"context"
	"database/sql"
	"strconv"
	"sync"
	"time"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/jobs_queries"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

// fakeTasksStorage and fakeJobsStorage are minimal in-memory implementations
// of storage.TasksStorage/storage.JobsStorage, following this repo's existing
// convention of hand-written test fakes (see verv_services/list_test.go)
// rather than generated mocks, for the small Querier interfaces here.

type fakeTasksStorage struct {
	mu     sync.Mutex
	nextID int64
	byID   map[int64]tasks_queries.VelezTask
}

func newFakeTasksStorage() *fakeTasksStorage {
	return &fakeTasksStorage{byID: make(map[int64]tasks_queries.VelezTask)}
}

func (f *fakeTasksStorage) CreateTask(_ context.Context, arg tasks_queries.CreateTaskParams) (tasks_queries.VelezTask, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, t := range f.byID {
		if t.EntityID == arg.EntityID && t.Action == arg.Action {
			return tasks_queries.VelezTask{}, sql.ErrNoRows
		}
	}

	f.nextID++
	task := tasks_queries.VelezTask{
		ID:        f.nextID,
		EntityID:  arg.EntityID,
		Action:    arg.Action,
		Status:    tasks_queries.VelezTaskStatusPENDING,
		Context:   arg.Context,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	f.byID[task.ID] = task

	return task, nil
}

func (f *fakeTasksStorage) GetTaskByEntityAction(_ context.Context, arg tasks_queries.GetTaskByEntityActionParams) (tasks_queries.VelezTask, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, t := range f.byID {
		if t.EntityID == arg.EntityID && t.Action == arg.Action {
			return t, nil
		}
	}

	return tasks_queries.VelezTask{}, sql.ErrNoRows
}

func (f *fakeTasksStorage) GetTaskById(_ context.Context, id int64) (tasks_queries.VelezTask, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	t, ok := f.byID[id]
	if !ok {
		return tasks_queries.VelezTask{}, sql.ErrNoRows
	}

	return t, nil
}

func (f *fakeTasksStorage) ClaimTask(_ context.Context, arg tasks_queries.ClaimTaskParams) (tasks_queries.VelezTask, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	staleThreshold := arg.ClaimedAt.Time

	for id, t := range f.byID {
		claimable := t.Status == tasks_queries.VelezTaskStatusPENDING ||
			(t.Status == tasks_queries.VelezTaskStatusRUNNING && t.ClaimedAt.Valid && t.ClaimedAt.Time.Before(staleThreshold))

		if !claimable {
			continue
		}

		t.Status = tasks_queries.VelezTaskStatusRUNNING
		t.ClaimedAt = sql.NullTime{Time: time.Now(), Valid: true}
		t.ClaimedBy = arg.ClaimedBy
		t.UpdatedAt = time.Now()
		f.byID[id] = t

		return t, nil
	}

	return tasks_queries.VelezTask{}, sql.ErrNoRows
}

func (f *fakeTasksStorage) UpdateTaskContext(_ context.Context, arg tasks_queries.UpdateTaskContextParams) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	t, ok := f.byID[arg.ID]
	if !ok {
		return sql.ErrNoRows
	}

	t.Context = arg.Context
	t.UpdatedAt = time.Now()
	f.byID[arg.ID] = t

	return nil
}

func (f *fakeTasksStorage) FinishTask(_ context.Context, arg tasks_queries.FinishTaskParams) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	t, ok := f.byID[arg.ID]
	if !ok {
		return sql.ErrNoRows
	}

	t.Status = arg.Status
	t.Error = arg.Error
	t.UpdatedAt = time.Now()
	f.byID[arg.ID] = t

	return nil
}

func (f *fakeTasksStorage) WithTx(_ *sql.Tx) *tasks_queries.Queries {
	return nil
}

func (f *fakeTasksStorage) get(id int64) tasks_queries.VelezTask {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.byID[id]
}

type fakeJobsStorage struct {
	mu   sync.Mutex
	rows map[string]jobs_queries.VelezJob
}

func newFakeJobsStorage() *fakeJobsStorage {
	return &fakeJobsStorage{rows: make(map[string]jobs_queries.VelezJob)}
}

func jobKey(taskID int64, jobName string) string {
	return strconv.FormatInt(taskID, 10) + "@" + jobName
}

func (f *fakeJobsStorage) GetJob(_ context.Context, arg jobs_queries.GetJobParams) (jobs_queries.VelezJob, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	row, ok := f.rows[jobKey(arg.TaskID, arg.JobName)]
	if !ok {
		return jobs_queries.VelezJob{}, sql.ErrNoRows
	}

	return row, nil
}

func (f *fakeJobsStorage) CreateRunningJob(_ context.Context, arg jobs_queries.CreateRunningJobParams) (jobs_queries.VelezJob, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	key := jobKey(arg.TaskID, arg.JobName)
	if _, ok := f.rows[key]; ok {
		return jobs_queries.VelezJob{}, sql.ErrNoRows
	}

	row := jobs_queries.VelezJob{
		TaskID:    arg.TaskID,
		JobName:   arg.JobName,
		Status:    jobs_queries.VelezJobStatusRUNNING,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	f.rows[key] = row

	return row, nil
}

func (f *fakeJobsStorage) FinishJob(_ context.Context, arg jobs_queries.FinishJobParams) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	key := jobKey(arg.TaskID, arg.JobName)
	row, ok := f.rows[key]
	if !ok {
		return sql.ErrNoRows
	}

	row.Status = arg.Status
	row.Error = arg.Error
	row.UpdatedAt = time.Now()
	f.rows[key] = row

	return nil
}

func (f *fakeJobsStorage) ListJobsByTask(_ context.Context, taskID int64) ([]jobs_queries.VelezJob, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	out := make([]jobs_queries.VelezJob, 0)
	for _, row := range f.rows {
		if row.TaskID == taskID {
			out = append(out, row)
		}
	}

	return out, nil
}

func (f *fakeJobsStorage) WithTx(_ *sql.Tx) *jobs_queries.Queries {
	return nil
}

func (f *fakeJobsStorage) seedDone(taskID int64, jobName string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.rows[jobKey(taskID, jobName)] = jobs_queries.VelezJob{
		TaskID:  taskID,
		JobName: jobName,
		Status:  jobs_queries.VelezJobStatusDONE,
	}
}

// fakeServicesStorage is a minimal in-memory implementation of
// storage.ServicesStorage for exercising create_service jobs without a
// database.
type fakeServicesStorage struct {
	mu       sync.Mutex
	upserted []string
}

func newFakeServicesStorage() *fakeServicesStorage {
	return &fakeServicesStorage{}
}

func (f *fakeServicesStorage) GetByName(_ context.Context, name string) (domain.Service, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, n := range f.upserted {
		if n == name {
			return domain.Service{ServiceBaseInfo: domain.ServiceBaseInfo{Name: name}}, nil
		}
	}

	return domain.Service{}, sql.ErrNoRows
}

func (f *fakeServicesStorage) UpsertService(_ context.Context, name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.upserted = append(f.upserted, name)

	return nil
}

func (f *fakeServicesStorage) List(_ context.Context, _ domain.ListServicesReq) (domain.ServiceList, error) {
	return domain.ServiceList{}, nil
}
