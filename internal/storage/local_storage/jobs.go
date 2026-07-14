package local_storage

import (
	"context"
	"database/sql"
	"strconv"
	"sync"
	"time"

	"go.vervstack.ru/Velez/internal/storage/postgres/generated/jobs_queries"
)

// jobs is the in-memory counterpart to tasks, see its comment for why this
// needs to be a real implementation rather than a no-op stub.
type jobs struct {
	mu   sync.Mutex
	rows map[string]jobs_queries.VelezJob
}

func newJobsStorage() *jobs {
	return &jobs{
		rows: make(map[string]jobs_queries.VelezJob),
	}
}

func jobKey(taskID int64, jobName string) string {
	return strconv.FormatInt(taskID, 10) + "@" + jobName
}

func (j *jobs) GetJob(_ context.Context, arg jobs_queries.GetJobParams) (jobs_queries.VelezJob, error) {
	j.mu.Lock()
	defer j.mu.Unlock()

	row, ok := j.rows[jobKey(arg.TaskID, arg.JobName)]
	if !ok {
		return jobs_queries.VelezJob{}, sql.ErrNoRows
	}

	return row, nil
}

func (j *jobs) CreateRunningJob(_ context.Context, arg jobs_queries.CreateRunningJobParams) (jobs_queries.VelezJob, error) {
	j.mu.Lock()
	defer j.mu.Unlock()

	key := jobKey(arg.TaskID, arg.JobName)
	if _, ok := j.rows[key]; ok {
		return jobs_queries.VelezJob{}, sql.ErrNoRows
	}

	now := time.Now()
	row := jobs_queries.VelezJob{
		TaskID:    arg.TaskID,
		JobName:   arg.JobName,
		Status:    jobs_queries.VelezJobStatusRUNNING,
		CreatedAt: now,
		UpdatedAt: now,
	}
	j.rows[key] = row

	return row, nil
}

func (j *jobs) FinishJob(_ context.Context, arg jobs_queries.FinishJobParams) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	key := jobKey(arg.TaskID, arg.JobName)
	row, ok := j.rows[key]
	if !ok {
		return sql.ErrNoRows
	}

	row.Status = arg.Status
	row.Error = arg.Error
	row.UpdatedAt = time.Now()
	j.rows[key] = row

	return nil
}

func (j *jobs) ListJobsByTask(_ context.Context, taskID int64) ([]jobs_queries.VelezJob, error) {
	j.mu.Lock()
	defer j.mu.Unlock()

	out := make([]jobs_queries.VelezJob, 0)
	for _, row := range j.rows {
		if row.TaskID == taskID {
			out = append(out, row)
		}
	}

	return out, nil
}

func (j *jobs) WithTx(_ *sql.Tx) *jobs_queries.Queries {
	return nil
}
