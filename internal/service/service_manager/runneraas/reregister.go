package runneraas

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/jobs"
)

// ReregisterRunner enqueues the reregister_runner task and returns
// immediately - mirrors CreateRunner. Callers watch progress through
// TasksApi.WatchTask(name, jobs.ReregisterRunnerAction).
func (s *RunneraasService) ReregisterRunner(ctx context.Context, name string) error {
	initialContext := &velez_api.ReregisterRunnerTaskPayload{
		Request: &velez_api.ReregisterRunner_Request{Name: name},
	}

	_, err := s.jobsEngine.Enqueue(ctx, name, jobs.ReregisterRunnerAction, initialContext)
	if err != nil {
		return rerrors.Wrap(err, "error enqueuing reregister runner task")
	}

	return nil
}
