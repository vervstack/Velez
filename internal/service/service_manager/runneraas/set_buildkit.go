package runneraas

import (
	"context"
	"strconv"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/jobs"
)

// SetRunnerBuildkit enqueues the set_runner_buildkit task and returns its id.
// Each call starts a new task once the previous one for the runner has
// finished, so BuildKit can be toggled repeatedly and every toggle stays in the
// task history. BuildKit is supported only for DinD-backed GitLab runners.
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

	return s.enqueueSetRunnerBuildkit(ctx, name, runner, isBuildkitEnabled, false)
}

// enqueueSetRunnerBuildkit starts the set_runner_buildkit task for an already
// loaded runner row. isWaitingForRedeploy holds the enable flow back until the
// recreated runner container is up on its DinD - see
// velez_api.SetRunnerBuildkitTaskPayload.is_waiting_for_redeploy.
func (s *RunneraasService) enqueueSetRunnerBuildkit(
	ctx context.Context, name string, runner domain.Runner, isBuildkitEnabled, isWaitingForRedeploy bool,
) (string, error) {
	provider := velez_api.RunnerProvider(velez_api.RunnerProvider_value[runner.Provider])

	err := validateBuildkitRunner(provider, runner.DindServiceId != 0)
	if err != nil {
		return "", rerrors.Wrap(err)
	}

	initialContext := &velez_api.SetRunnerBuildkitTaskPayload{
		Request: &velez_api.SetRunnerBuildkit_Request{
			Name:              name,
			IsBuildkitEnabled: isBuildkitEnabled,
		},
		IsWaitingForRedeploy: isWaitingForRedeploy,
	}

	task, err := s.jobsEngine.Enqueue(ctx, name, jobs.SetRunnerBuildkitAction, initialContext)
	if err != nil {
		return "", rerrors.Wrap(err, "error enqueuing set runner buildkit task")
	}

	return strconv.FormatInt(task.ID, 10), nil
}
