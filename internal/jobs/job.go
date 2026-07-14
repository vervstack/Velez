package jobs

import "context"

// Job is a tiny, atomic, reusable unit of execution inside a Task.
// Unlike internal/pipelines/steps.Step, a Job's completion is durably
// checkpointed, so re-running the same Job for the same task is safe.
type Job interface {
	Do(ctx context.Context) error
}

// RollbackableJob lets a Job undo its own side effects if a later Job
// in the same task attempt fails.
type RollbackableJob interface {
	Rollback(ctx context.Context) error
}

// NamedJob pairs a Job with the name it's checkpointed under.
type NamedJob struct {
	Name string
	Job  Job
}

// TaskContext is a task's persisted business-data payload (a proto message
// pointer in practice). Jobs never depend on the concrete type directly -
// they declare narrow Getter/Setter accessor interfaces instead, so the
// same Job can be reused across any TaskContext that satisfies them.
type TaskContext interface{}

// TaskHandler builds the ordered Job list for one task action.
type TaskHandler interface {
	// Action identifies the task, e.g. "create_smerd".
	Action() string
	// NewContext returns a zero-value, JSON-unmarshalable TaskContext
	// (a pointer), used both for a brand-new task and to restore a
	// resumed one from its persisted context.
	NewContext() TaskContext
	// BuildJobs returns the ordered jobs for taskCtx. Jobs close over
	// taskCtx via accessor interfaces to read/write shared task data.
	BuildJobs(taskCtx TaskContext) []NamedJob
}
