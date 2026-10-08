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
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	defaultClaimLease = 2 * time.Minute
	rollbackTimeout   = 30 * time.Second

	claimRenewalsPerLease = 4
)

// taskWorker generalizes internal/workers/deploy_watcher.go's ticker-driven
// polling into a claim-any-registered-action loop backed by SELECT ... FOR
// UPDATE SKIP LOCKED, with reclaim of tasks a worker died while holding.
//
// It satisfies internal/workers.Worker structurally, but NewTaskWorker
// returns the concrete type rather than that interface: deploy_watcher.go now
// enqueues jobs-engine tasks, so internal/workers imports internal/jobs and
// the reverse edge would be an import cycle.
type taskWorker struct {
	tasksStorage storage.TasksStorage
	jobsStorage  storage.JobsStorage
	registry     *Registry

	workerID    string
	lease       time.Duration
	concurrency int

	starter  sync.Once
	stopOnce sync.Once
	ticker   *time.Ticker
	done     chan struct{}

	activeMu sync.Mutex
	active   map[int64]*activeTask
}

// activeTask tracks what a taskWorker goroutine is doing right now, so Stop
// can report exactly what shutdown is not waiting for instead of silently
// abandoning it.
type activeTask struct {
	action  string
	jobName string
}

// NewTaskWorker's concurrency is how many goroutines independently poll and
// claim tasks. ClaimTask is safe for concurrent callers by design (SELECT
// ... FOR UPDATE SKIP LOCKED in Postgres, a mutex-guarded atomic claim in
// local_storage - see tasks.go's doc comment), and a single-goroutine worker
// self-deadlocks the moment any job blocks waiting on another task to reach
// a terminal status (e.g. jobs.waitForRegistryDeployJob watching the
// create_smerd task deploy_watcher.go dispatches): that other task can never
// be claimed because the only worker is busy waiting for it.
func NewTaskWorker(
	tasksStorage storage.TasksStorage,
	jobsStorage storage.JobsStorage,
	registry *Registry,
	workerID string,
	interval time.Duration,
	concurrency int,
) *taskWorker {
	return &taskWorker{
		tasksStorage: tasksStorage,
		jobsStorage:  jobsStorage,
		registry:     registry,

		workerID:    workerID,
		lease:       defaultClaimLease,
		concurrency: concurrency,

		ticker: time.NewTicker(interval),
		done:   make(chan struct{}),
		active: make(map[int64]*activeTask),
	}
}

func (w *taskWorker) Start(ctx context.Context) {
	w.starter.Do(func() {
		var wg sync.WaitGroup

		for range w.concurrency {
			wg.Add(1)

			go func() {
				defer wg.Done()

				w.pollLoop(ctx)
			}()
		}

		wg.Wait()
	})
}

func (w *taskWorker) Stop() error {
	w.stopOnce.Do(func() {
		w.ticker.Stop()
		close(w.done)
		w.logActiveTasks()
	})

	return nil
}

// logActiveTasks reports every task this worker was mid-execution on at the
// moment Stop was called. Stop only halts new claims - it never waits for or
// cancels a task already running - so this is the only visibility into what
// shutdown is abandoning.
func (w *taskWorker) logActiveTasks() {
	w.activeMu.Lock()
	defer w.activeMu.Unlock()

	for taskID, at := range w.active {
		log.Warn().
			Int64("task_id", taskID).
			Str("action", at.action).
			Str("job", at.jobName).
			Msg("task still running at shutdown, not waited for")
	}
}

func (w *taskWorker) setActiveTask(taskID int64, action string) {
	w.activeMu.Lock()
	defer w.activeMu.Unlock()

	w.active[taskID] = &activeTask{action: action}
}

func (w *taskWorker) setActiveJob(taskID int64, jobName string) {
	w.activeMu.Lock()
	defer w.activeMu.Unlock()

	at, ok := w.active[taskID]
	if !ok {
		return
	}

	at.jobName = jobName
}

func (w *taskWorker) clearActiveTask(taskID int64) {
	w.activeMu.Lock()
	defer w.activeMu.Unlock()

	delete(w.active, taskID)
}

func (w *taskWorker) pollLoop(ctx context.Context) {
	for {
		select {
		case <-w.done:
			return
		case <-w.ticker.C:
			w.processOne(ctx)
		}
	}
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
			Err(rerrors.Wrap(err, "error running task")).
			Int64("task_id", task.ID).
			Str("action", task.Action).
			Msg("error running task")
	}
}

func (w *taskWorker) run(ctx context.Context, task tasks_queries.VelezTask) error {
	w.setActiveTask(task.ID, task.Action)
	defer w.clearActiveTask(task.ID)

	stopHeartbeat := w.keepClaim(ctx, task.ID)
	defer stopHeartbeat()

	handler, ok := w.registry.Get(task.Action)
	if !ok {
		return w.failTask(ctx, task.ID, rerrors.Wrap(user_errors.ErrNoHandlerRegisteredForAction, task.Action))
	}

	taskCtx := handler.NewContext()
	if task.Context.Valid {
		err := json.Unmarshal(task.Context.RawMessage, taskCtx)
		if err != nil {
			return w.failTask(ctx, task.ID, rerrors.Wrap(err, "error unmarshaling task context"))
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

// keepClaim renews the claim on a running task until the returned stop is
// called. A job can run longer than the lease (a cold image pull); without the
// renewal another worker goroutine would reclaim the task as stale and run its
// jobs a second time in parallel.
func (w *taskWorker) keepClaim(ctx context.Context, taskID int64) func() {
	stop := make(chan struct{})
	finished := make(chan struct{})

	renewParams := tasks_queries.RenewTaskClaimParams{
		ID:        taskID,
		ClaimedBy: sql.NullString{String: w.workerID, Valid: true},
	}

	go func() {
		defer close(finished)

		ticker := time.NewTicker(w.lease / claimRenewalsPerLease)
		defer ticker.Stop()

		for {
			select {
			case <-stop:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				err := w.tasksStorage.RenewTaskClaim(ctx, renewParams)
				if err != nil {
					log.Warn().Err(err).Int64("task_id", taskID).Msg("error renewing task claim")
				}
			}
		}
	}()

	return func() {
		close(stop)
		<-finished
	}
}

// failTask marks a task FAILED before any job ever ran for it (no handler
// registered, or the persisted context couldn't be unmarshaled) - without
// this, a task that fails here would stay claimed at RUNNING forever, only
// reclaimable after the stale lease expires, and would fail identically on
// every retry.
func (w *taskWorker) failTask(ctx context.Context, taskID int64, taskErr error) error {
	finishErr := w.tasksStorage.FinishTask(ctx, tasks_queries.FinishTaskParams{
		ID:     taskID,
		Status: tasks_queries.VelezTaskStatusFAILED,
		Error:  sql.NullString{String: taskErr.Error(), Valid: true},
	})
	if finishErr != nil {
		return rerrors.Join(taskErr, finishErr)
	}

	return taskErr
}

func (w *taskWorker) runJobs(ctx context.Context, taskID int64, taskCtx TaskContext, namedJobs []NamedJob) error {
	failedIdx := -1

	var runErr error

	for i, nj := range namedJobs {
		w.setActiveJob(taskID, nj.Name)

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

	rollbackCtx, cancel := context.WithTimeout(context.Background(), rollbackTimeout)
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

	return rerrors.Wrapf(runErr, "job %q failed", namedJobs[failedIdx].Name)
}
