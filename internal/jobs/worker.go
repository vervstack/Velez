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

	"go.vervstack.ru/Velez/internal/service/user_errors"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

const (
	defaultClaimLease = 2 * time.Minute
	rollbackTimeout   = 30 * time.Second

	claimRenewalsPerLease = 4

	drainCheckInterval = 100 * time.Millisecond
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

	drainTimeout    time.Duration
	finishedTaskTtl time.Duration

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

// SetDrainTimeout bounds how long Stop waits for in-flight tasks to finish.
// Zero (the default) keeps the old behaviour: Stop does not wait.
func (w *taskWorker) SetDrainTimeout(timeout time.Duration) {
	w.drainTimeout = timeout
}

// SetFinishedTaskTtl enables evicting DONE tasks that finished longer ago than
// ttl from the task store, checked on every fallback tick. Zero (the default)
// disables eviction.
func (w *taskWorker) SetFinishedTaskTtl(ttl time.Duration) {
	w.finishedTaskTtl = ttl
}

func (w *taskWorker) Start(ctx context.Context) {
	w.starter.Do(func() {
		var wg sync.WaitGroup

		for range w.concurrency {
			wg.Add(1)

			wake := w.tasksStorage.SubscribeTaskCreated(ctx)

			go func() {
				defer wg.Done()

				w.pollLoop(ctx, wake)
			}()
		}

		wg.Wait()
	})
}

// Stop halts new claims and waits up to the drain timeout for in-flight tasks
// to finish; whatever is still running after that is logged and abandoned.
// Repeated calls return immediately.
func (w *taskWorker) Stop() error {
	w.stopOnce.Do(func() {
		w.ticker.Stop()
		close(w.done)
		w.waitForActiveTasks()
		w.logActiveTasks()
	})

	return nil
}

func (w *taskWorker) waitForActiveTasks() {
	deadline := time.Now().Add(w.drainTimeout)

	for w.activeCount() > 0 && time.Now().Before(deadline) {
		time.Sleep(drainCheckInterval)
	}
}

func (w *taskWorker) activeCount() int {
	w.activeMu.Lock()
	defer w.activeMu.Unlock()

	return len(w.active)
}

// logActiveTasks reports every task this worker is still executing once the
// drain gave up, so the log shows exactly what shutdown abandoned.
func (w *taskWorker) logActiveTasks() {
	w.activeMu.Lock()
	defer w.activeMu.Unlock()

	for taskID, at := range w.active {
		log.Warn().
			Int64("task_id", taskID).
			Str("action", at.action).
			Str("job", at.jobName).
			Msg("task still running after drain timeout, abandoned")
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

// pollLoop wakes on a task-created signal and, as a fallback, on the ticker: a
// NOTIFY sent while the listener connection was down is lost, and stale
// leases left by a dead worker are only ever found by polling.
func (w *taskWorker) pollLoop(ctx context.Context, wake <-chan struct{}) {
	w.drainQueue(ctx)

	for {
		select {
		case <-w.done:
			return
		case <-wake:
		case <-w.ticker.C:
			w.evictFinishedTasks(ctx)
		}

		w.drainQueue(ctx)
	}
}

func (w *taskWorker) evictFinishedTasks(ctx context.Context) {
	if w.finishedTaskTtl <= 0 {
		return
	}

	evictedIds, err := w.tasksStorage.EvictFinished(ctx, time.Now().Add(-w.finishedTaskTtl))
	if err != nil {
		log.Warn().Err(err).Msg("error evicting finished tasks")

		return
	}

	for _, taskID := range evictedIds {
		err = w.jobsStorage.DeleteJobsByTask(ctx, taskID)
		if err != nil {
			log.Warn().Err(err).Int64("task_id", taskID).Msg("error deleting jobs of evicted task")
		}
	}
}

func (w *taskWorker) drainQueue(ctx context.Context) {
	for w.processOne(ctx) {
		select {
		case <-w.done:
			return
		default:
		}
	}
}

// processOne reports whether it claimed a task.
func (w *taskWorker) processOne(ctx context.Context) bool {
	select {
	case <-w.done:
		return false
	default:
	}

	task, err := w.tasksStorage.ClaimTask(ctx, tasks_queries.ClaimTaskParams{
		ClaimedBy: sql.NullString{String: w.workerID, Valid: true},
		ClaimedAt: sql.NullTime{Time: time.Now().Add(-w.lease), Valid: true},
	})
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Error().Err(err).Msg("error claiming task")
		}

		return false
	}

	err = w.run(WithInternalEnqueue(ctx), task)
	if err != nil {
		log.Error().
			Err(rerrors.Wrap(err, "error running task")).
			Int64("task_id", task.ID).
			Str("action", task.Action).
			Msg("error running task")
	}

	return true
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

	if runErr != nil {
		return rerrors.Wrap(runErr)
	}

	return nil
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

	if taskErr != nil {
		return rerrors.Wrap(taskErr)
	}

	return nil
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
