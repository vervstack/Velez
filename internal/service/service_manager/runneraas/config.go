package runneraas

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service/service_manager/runneraas/providers"
	"go.vervstack.ru/Velez/internal/user_errors"
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
		Concurrent:          runner.Concurrent,
	}

	return config, nil
}

// UpdateRunnerConfig persists only the Valid fields of req onto the runner's
// row - see domain.UpdateRunnerConfigReq's doc comment. Concurrent is the one
// field applied to the running container immediately; for every other one the
// returned domain.UpdateRunnerConfigResult tells the caller which follow-up
// action does.
func (s *RunneraasService) UpdateRunnerConfig(
	ctx context.Context, req domain.UpdateRunnerConfigReq,
) (domain.UpdateRunnerConfigResult, error) {
	if req.DockerSocketAddress.Valid {
		err := validateDockerSocketAddress(req.DockerSocketAddress.Value)
		if err != nil {
			return domain.UpdateRunnerConfigResult{}, rerrors.Wrap(err, "error validating docker socket address")
		}
	}

	if req.Concurrent.Valid && req.Concurrent.Value < 1 {
		return domain.UpdateRunnerConfigResult{}, rerrors.Wrap(user_errors.ErrRunnerConcurrentInvalid)
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
		Concurrent:          runner.Concurrent,
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

	if req.Concurrent.Valid {
		err = s.applyConcurrent(ctx, svc.Env, runner.Provider, req.Name, req.Concurrent.Value)
		if err != nil {
			return domain.UpdateRunnerConfigResult{}, rerrors.Wrap(err, "error applying concurrent")
		}

		upsertReq.Concurrent = req.Concurrent.Value
	}

	_, err = s.dataStorage.Runners().UpsertRunner(ctx, upsertReq)
	if err != nil {
		return domain.UpdateRunnerConfigResult{}, rerrors.Wrap(err, "error upserting runner config")
	}

	return result, nil
}

func (s *RunneraasService) applyConcurrent(
	ctx context.Context, environment, provider, name string, concurrent int32,
) error {
	providerEnum := velez_api.RunnerProvider(velez_api.RunnerProvider_value[provider])

	runnerProvider, err := providers.For(providerEnum)
	if err != nil {
		return rerrors.Wrap(err, "error resolving runner provider")
	}

	containerRuntime, err := s.runtimes.Runtime(ctx, environment)
	if err != nil {
		return rerrors.Wrap(err, "error resolving container runtime")
	}

	err = runnerProvider.ApplyConcurrent(ctx, containerRuntime, name, concurrent)
	if err != nil {
		return rerrors.Wrap(err, "error applying provider concurrent")
	}

	return nil
}
