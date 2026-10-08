package runneraas

import (
	"context"
	"strconv"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/jobs"
)

// SetRunnerBuildkit enqueues the set_runner_buildkit task and returns its id.
// A finished earlier task for the same runner is replaced, so the sidecar can
// be toggled repeatedly.
func (s *RunneraasService) SetRunnerBuildkit(
	ctx context.Context, name string, isBuildkitEnabled bool,
) (string, error) {
	svc, err := s.dataStorage.Services().GetByName(ctx, name)
	if err != nil {
		return "", rerrors.Wrap(err, "error getting runner service")
	}

	_, err = s.dataStorage.Runners().GetRunnerByServiceID(ctx, svc.ID)
	if err != nil {
		return "", rerrors.Wrap(err, "error getting runner row")
	}

	initialContext := &velez_api.SetRunnerBuildkitTaskPayload{
		Request: &velez_api.SetRunnerBuildkit_Request{
			Name:              name,
			IsBuildkitEnabled: isBuildkitEnabled,
		},
	}

	task, err := s.jobsEngine.EnqueueReplacing(ctx, name, jobs.SetRunnerBuildkitAction, initialContext)
	if err != nil {
		return "", rerrors.Wrap(err, "error enqueuing set runner buildkit task")
	}

	return strconv.FormatInt(task.ID, 10), nil
}
