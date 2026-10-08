package runneraas

import (
	"context"
	"strconv"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/jobs"
)

// SetRunnerBuildkit enqueues the set_runner_buildkit task and returns its id.
// Each call starts a new task once the previous one for the runner has
// finished, so the sidecar can be toggled repeatedly and every toggle stays in
// the task history. BuildKit is supported only for runners whose jobs use the
// node's own docker socket.
func (s *RunneraasService) SetRunnerBuildkit(
	ctx context.Context, name string, isBuildkitEnabled bool,
) (string, error) {
	svc, err := s.dataStorage.Services().GetByName(ctx, name)
	if err != nil {
		return "", rerrors.Wrap(err, "error getting runner service")
	}

	runner, err := s.dataStorage.Runners().GetRunnerByServiceID(ctx, svc.ID)
	if err != nil {
		return "", rerrors.Wrap(err, "error getting runner row")
	}

	err = validateBuildkitDockerSource(runner.DindServiceId != 0, runner.DockerSocketAddress)
	if err != nil {
		return "", rerrors.Wrap(err)
	}

	initialContext := &velez_api.SetRunnerBuildkitTaskPayload{
		Request: &velez_api.SetRunnerBuildkit_Request{
			Name:              name,
			IsBuildkitEnabled: isBuildkitEnabled,
		},
	}

	task, err := s.jobsEngine.Enqueue(ctx, name, jobs.SetRunnerBuildkitAction, initialContext)
	if err != nil {
		return "", rerrors.Wrap(err, "error enqueuing set runner buildkit task")
	}

	return strconv.FormatInt(task.ID, 10), nil
}
