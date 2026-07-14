package jobs

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/sqlc-dev/pqtype"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/jobs_queries"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

// checkpointedJob gates a Job's execution behind a durable (task_id, job_name)
// row: DONE is skipped, FAILED short-circuits with the stored error, and a
// successful run persists the task's current context so a crash-resume can
// restore exactly what earlier jobs produced.
type checkpointedJob struct {
	jobsStorage  storage.JobsStorage
	tasksStorage storage.TasksStorage

	taskID  int64
	jobName string
	taskCtx TaskContext

	inner Job
}

func Checkpoint(
	jobsStorage storage.JobsStorage,
	tasksStorage storage.TasksStorage,
	taskID int64,
	jobName string,
	taskCtx TaskContext,
	inner Job,
) Job {
	return &checkpointedJob{
		jobsStorage:  jobsStorage,
		tasksStorage: tasksStorage,
		taskID:       taskID,
		jobName:      jobName,
		taskCtx:      taskCtx,
		inner:        inner,
	}
}

func (c *checkpointedJob) Do(ctx context.Context) error {
	existing, err := c.jobsStorage.GetJob(ctx, jobs_queries.GetJobParams{
		TaskID:  c.taskID,
		JobName: c.jobName,
	})

	switch {
	case err == nil:
		switch existing.Status {
		case jobs_queries.VelezJobStatusDONE:
			return nil
		case jobs_queries.VelezJobStatusFAILED:
			errMsg := c.jobName + " previously failed"
			if existing.Error.Valid {
				errMsg = existing.Error.String
			}
			return rerrors.New(errMsg)
		default:
			// RUNNING left over from a worker that crashed mid-job - re-run it.
		}
	case errors.Is(err, sql.ErrNoRows):
		_, err = c.jobsStorage.CreateRunningJob(ctx, jobs_queries.CreateRunningJobParams{
			TaskID:  c.taskID,
			JobName: c.jobName,
		})
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return rerrors.Wrap(err, "error creating job checkpoint")
		}
	default:
		return rerrors.Wrap(err, "error reading job checkpoint")
	}

	doErr := c.inner.Do(ctx)

	finishParams := jobs_queries.FinishJobParams{
		TaskID:  c.taskID,
		JobName: c.jobName,
		Status:  jobs_queries.VelezJobStatusDONE,
	}
	if doErr != nil {
		finishParams.Status = jobs_queries.VelezJobStatusFAILED
		finishParams.Error = sql.NullString{String: doErr.Error(), Valid: true}
	}

	finishErr := c.jobsStorage.FinishJob(ctx, finishParams)
	if finishErr != nil {
		return rerrors.Wrap(rerrors.Join(doErr, finishErr), "error finishing job checkpoint")
	}

	if doErr != nil {
		return doErr
	}

	return c.persistContext(ctx)
}

func (c *checkpointedJob) persistContext(ctx context.Context) error {
	contextJSON, err := json.Marshal(c.taskCtx)
	if err != nil {
		return rerrors.Wrap(err, "error marshaling task context")
	}

	err = c.tasksStorage.UpdateTaskContext(ctx, tasks_queries.UpdateTaskContextParams{
		ID:      c.taskID,
		Context: pqtype.NullRawMessage{RawMessage: contextJSON, Valid: true},
	})
	if err != nil {
		return rerrors.Wrap(err, "error persisting task context")
	}

	return nil
}

func (c *checkpointedJob) Rollback(ctx context.Context) error {
	rb, ok := c.inner.(RollbackableJob)
	if !ok {
		return nil
	}

	return rb.Rollback(ctx)
}
