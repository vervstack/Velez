package jobs

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sqlc-dev/pqtype"
	"go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/jobs_queries"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

const (
	defaultWatchPollInterval = time.Second
)

// Engine lets callers enqueue durable, resumable tasks and observe their
// progress, without needing to know about the worker that executes them.
type Engine interface {
	// Enqueue creates a task for (entityID, action) if one doesn't already
	// exist. If it does, the existing task is returned instead - the caller
	// inspects Status: in-flight means attach via Watch, FAILED means
	// surface the prior failure.
	Enqueue(ctx context.Context, entityID, action string, initialContext any) (tasks_queries.VelezTask, error)
	// Watch streams task status changes for (entityID, action) until the
	// task reaches a terminal status (DONE/FAILED), then closes the channel.
	Watch(ctx context.Context, entityID, action string) <-chan tasks_queries.VelezTask

	// ListJobs returns the ordered per-job status breakdown for task, merging the
	// action's full BuildJobs-defined job list (so not-yet-started jobs appear as
	// pending) with whatever's actually been persisted so far. Returns nil, nil if
	// the task's action has no registered handler (caller should degrade to
	// showing only the overall task status).
	ListJobs(ctx context.Context, task tasks_queries.VelezTask) ([]JobStatus, error)

	// SetRegistry attaches the job registry once it's been built (it depends on
	// c.Services, which doesn't exist yet when the engine itself is constructed
	// in internal/app/custom.go's InitServiceLayer - see that file's wiring).
	SetRegistry(registry *Registry)
}

// JobStatus is one job's name paired with its persisted execution status,
// suitable for exposing a task's per-job progress breakdown to callers.
type JobStatus struct {
	Name string
	// Status is the zero value "" when the job hasn't been persisted yet
	// (not started / pending).
	Status jobs_queries.VelezJobStatus
}

type engine struct {
	tasksStorage storage.TasksStorage
	jobsStorage  storage.JobsStorage
	pollInterval time.Duration

	registry *Registry
}

func NewEngine(tasksStorage storage.TasksStorage, jobsStorage storage.JobsStorage) Engine {
	return &engine{
		tasksStorage: tasksStorage,
		jobsStorage:  jobsStorage,
		pollInterval: defaultWatchPollInterval,
	}
}

func (e *engine) SetRegistry(registry *Registry) {
	e.registry = registry
}

func (e *engine) ListJobs(ctx context.Context, task tasks_queries.VelezTask) ([]JobStatus, error) {
	if e.registry == nil {
		return nil, nil
	}

	handler, ok := e.registry.Get(task.Action)
	if !ok {
		return nil, nil
	}

	taskCtx := handler.NewContext()
	if task.Context.Valid {
		err := json.Unmarshal(task.Context.RawMessage, taskCtx)
		if err != nil {
			return nil, rerrors.Wrap(err, "error unmarshaling task context")
		}
	}

	namedJobs := handler.BuildJobs(taskCtx)

	rows, err := e.jobsStorage.ListJobsByTask(ctx, task.ID)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing jobs for task")
	}

	statusByName := make(map[string]jobs_queries.VelezJobStatus, len(rows))
	for _, row := range rows {
		statusByName[row.JobName] = row.Status
	}

	out := make([]JobStatus, len(namedJobs))
	for i, nj := range namedJobs {
		out[i] = JobStatus{
			Name:   nj.Name,
			Status: statusByName[nj.Name],
		}
	}

	return out, nil
}

func (e *engine) Enqueue(
	ctx context.Context, entityID, action string, initialContext any,
) (tasks_queries.VelezTask, error) {
	contextJSON, err := json.Marshal(initialContext)
	if err != nil {
		return tasks_queries.VelezTask{}, rerrors.Wrap(err, "error marshaling initial context")
	}

	params := tasks_queries.CreateTaskParams{
		EntityID: entityID,
		Action:   action,
		Context:  pqtype.NullRawMessage{RawMessage: contextJSON, Valid: true},
	}

	task, err := e.tasksStorage.CreateTask(ctx, params)
	if err == nil {
		return task, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return tasks_queries.VelezTask{}, rerrors.Wrap(err, "error creating task")
	}

	task, err = e.tasksStorage.GetTaskByEntityAction(ctx, tasks_queries.GetTaskByEntityActionParams{
		EntityID: entityID,
		Action:   action,
	})
	if err != nil {
		return tasks_queries.VelezTask{}, rerrors.Wrap(err, "error fetching existing task")
	}

	return task, nil
}

func (e *engine) Watch(ctx context.Context, entityID, action string) <-chan tasks_queries.VelezTask {
	ch := make(chan tasks_queries.VelezTask)

	go func() {
		defer close(ch)

		ticker := time.NewTicker(e.pollInterval)
		defer ticker.Stop()

		var lastStatus tasks_queries.VelezTaskStatus

		for {
			task, err := e.tasksStorage.GetTaskByEntityAction(ctx, tasks_queries.GetTaskByEntityActionParams{
				EntityID: entityID,
				Action:   action,
			})

			switch {
			case err == nil:
				if task.Status != lastStatus {
					lastStatus = task.Status
					select {
					case ch <- task:
					case <-ctx.Done():
						return
					}
				}

				if task.Status == tasks_queries.VelezTaskStatusDONE || task.Status == tasks_queries.VelezTaskStatusFAILED {
					return
				}
			case errors.Is(err, sql.ErrNoRows):
				// task not created yet - keep polling
			default:
				log.Error().Err(err).Str("entity_id", entityID).Str("action", action).Msg("error watching task")

				return
			}

			select {
			case <-ticker.C:
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch
}
