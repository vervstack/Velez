package runneraas

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
)

// GetRunnerConfig reads the runner's currently stored provider config - see
// domain.RunnerConfig's doc comment on why that can differ from what the
// running container was deployed/registered with.
func (s *RunneraasService) GetRunnerConfig(ctx context.Context, name string) (domain.RunnerConfig, error) {
	svc, err := s.dataStorage.Services().GetByName(ctx, name)
	if err != nil {
		return domain.RunnerConfig{}, rerrors.Wrap(err, "error getting runner service")
	}

	runner, err := s.dataStorage.Runners().GetRunnerByServiceID(ctx, svc.ID)
	if err != nil {
		return domain.RunnerConfig{}, rerrors.Wrap(err, "error getting runner row")
	}

	config := domain.RunnerConfig{
		BaseUrl:             runner.BaseUrl,
		DockerImage:         runner.DockerImage,
		DockerSocketAddress: runner.DockerSocketAddress,
	}

	return config, nil
}

// UpdateRunnerConfig persists only the Valid fields of req onto the runner's
// row - see domain.UpdateRunnerConfigReq's doc comment. Never applies them
// to the running container; the returned domain.UpdateRunnerConfigResult
// tells the caller which follow-up action does.
func (s *RunneraasService) UpdateRunnerConfig(
	ctx context.Context, req domain.UpdateRunnerConfigReq,
) (domain.UpdateRunnerConfigResult, error) {
	if req.DockerSocketAddress.Valid {
		err := validateDockerSocketAddress(req.DockerSocketAddress.Value)
		if err != nil {
			return domain.UpdateRunnerConfigResult{}, rerrors.Wrap(err, "error validating docker socket address")
		}
	}

	svc, err := s.dataStorage.Services().GetByName(ctx, req.Name)
	if err != nil {
		return domain.UpdateRunnerConfigResult{}, rerrors.Wrap(err, "error getting runner service")
	}

	runner, err := s.dataStorage.Runners().GetRunnerByServiceID(ctx, svc.ID)
	if err != nil {
		return domain.UpdateRunnerConfigResult{}, rerrors.Wrap(err, "error getting runner row")
	}

	upsertReq := domain.UpsertRunnerReq{
		ServiceID:           svc.ID,
		Provider:            runner.Provider,
		Scope:               runner.Scope,
		Target:              runner.Target,
		Labels:              runner.Labels,
		SecretRef:           runner.SecretRef,
		BaseUrl:             runner.BaseUrl,
		DockerImage:         runner.DockerImage,
		DockerSocketAddress: runner.DockerSocketAddress,
	}

	var result domain.UpdateRunnerConfigResult

	if req.BaseUrl.Valid {
		upsertReq.BaseUrl = req.BaseUrl.Value
		result.RequiresReregister = true
	}

	if req.DockerImage.Valid {
		upsertReq.DockerImage = req.DockerImage.Value
		result.RequiresReregister = true
	}

	if req.DockerSocketAddress.Valid {
		upsertReq.DockerSocketAddress = req.DockerSocketAddress.Value
		result.RequiresRedeploy = true
	}

	_, err = s.dataStorage.Runners().UpsertRunner(ctx, upsertReq)
	if err != nil {
		return domain.UpdateRunnerConfigResult{}, rerrors.Wrap(err, "error upserting runner config")
	}

	return result, nil
}
