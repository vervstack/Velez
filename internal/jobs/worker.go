package jobs

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
	"go.vervstack.ru/Velez/internal/workers"
)

const defaultClaimLease = 2 * time.Minute

// taskWorker generalizes internal/workers/deploy_watcher.go's ticker-driven
// polling into a claim-any-registered-action loop backed by SELECT ... FOR
// UPDATE SKIP LOCKED, with reclaim of tasks a worker died while holding.
type taskWorker struct {
	tasksStorage storage.TasksStorage
	jobsStorage  storage.JobsStorage
	registry     *Registry

	workerID string
	lease    time.Duration

	starter  sync.Once
	stopOnce sync.Once
	ticker   *time.Ticker
	done     chan struct{}
}

func NewTaskWorker(
	tasksStorage storage.TasksStorage,
	jobsStorage storage.JobsStorage,
	registry *Registry,
	workerID string,
	interval time.Duration,
) workers.Worker {
	return &taskWorker{
		tasksStorage: tasksStorage,
		jobsStorage:  jobsStorage,
		registry:     registry,

		workerID: workerID,
		lease:    defaultClaimLease,

		ticker: time.NewTicker(interval),
		done:   make(chan struct{}),
	}
}

func (w *taskWorker) Start(ctx context.Context) {
	w.starter.Do(func() {
		for {
			select {
			case <-w.done:
				return
			case <-w.ticker.C:
				w.processOne(ctx)
			}
		}
	})
}

func (w *taskWorker) Stop() error {
	w.stopOnce.Do(func() {
		w.ticker.Stop()
		close(w.done)
	})
	return nil
}

func (w *taskWorker) processOne(ctx context.Context) {
	task, err := w.tasksStorage.ClaimTask(ctx, tasks_queries.ClaimTaskParams{
		ClaimedBy: sql.NullString{String: w.workerID, Valid: true},
		ClaimedAt: sql.NullTime{Time: time.Now().Add(-w.lease), Valid: true},
	})
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Error().Err(err).Msg("error claiming task")
		}
		return
	}

	err = w.run(ctx, task)
	if err != nil {
		log.Error().
			Err(rerrors.Wrap(err)).
			Int64("task_id", task.ID).
			Str("action", task.Action).
			Msg("error running task")
	}
}

func (w *taskWorker) run(ctx context.Context, task tasks_queries.VelezTask) error {
	handler, ok := w.registry.Get(task.Action)
	if !ok {
		notFoundErr := rerrors.New("no handler registered for action " + task.Action)

		finishErr := w.tasksStorage.FinishTask(ctx, tasks_queries.FinishTaskParams{
			ID:     task.ID,
			Status: tasks_queries.VelezTaskStatusFAILED,
			Error:  sql.NullString{String: notFoundErr.Error(), Valid: true},
		})
		if finishErr != nil {
			return rerrors.Join(notFoundErr, finishErr)
		}

		return notFoundErr
	}

	taskCtx := handler.NewContext()
	if task.Context.Valid {
		err := json.Unmarshal(task.Context.RawMessage, taskCtx)
		if err != nil {
			return rerrors.Wrap(err, "error unmarshaling task context")
		}
	}

	namedJobs := handler.BuildJobs(taskCtx)

	runErr := w.runJobs(ctx, task.ID, taskCtx, namedJobs)

	finishParams := tasks_queries.FinishTaskParams{
		ID:     task.ID,
		Status: tasks_queries.VelezTaskStatusDONE,
	}
	if runErr != nil {
		finishParams.Status = tasks_queries.VelezTaskStatusFAILED
		finishParams.Error = sql.NullString{String: runErr.Error(), Valid: true}
	}

	err := w.tasksStorage.FinishTask(ctx, finishParams)
	if err != nil {
		return rerrors.Join(runErr, rerrors.Wrap(err, "error finishing task"))
	}

	return runErr
}

func (w *taskWorker) runJobs(ctx context.Context, taskID int64, taskCtx TaskContext, namedJobs []NamedJob) error {
	failedIdx := -1
	var runErr error

	for i, nj := range namedJobs {
		checkpointed := Checkpoint(w.jobsStorage, w.tasksStorage, taskID, nj.Name, taskCtx, nj.Job)

		runErr = checkpointed.Do(ctx)
		if runErr != nil {
			failedIdx = i
			break
		}
	}

	if runErr == nil {
		return nil
	}

	rollbackCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for i := failedIdx; i >= 0; i-- {
		rb, ok := namedJobs[i].Job.(RollbackableJob)
		if !ok {
			continue
		}

		rbErr := rb.Rollback(rollbackCtx)
		if rbErr != nil {
			log.Error().
				Err(rbErr).
				Str("job", namedJobs[i].Name).
				Msg("error rolling back job")
		}
	}

	return runErr
}
