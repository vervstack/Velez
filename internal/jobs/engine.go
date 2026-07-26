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
}

type engine struct {
	tasksStorage storage.TasksStorage
	pollInterval time.Duration
}

func NewEngine(tasksStorage storage.TasksStorage) Engine {
	return &engine{
		tasksStorage: tasksStorage,
		pollInterval: defaultWatchPollInterval,
	}
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
