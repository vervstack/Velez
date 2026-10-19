package runneraas

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/service/user_errors"
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

	if runner.DindServiceId != 0 {
		config.DindName, err = s.dindName(ctx, runner.DindServiceId)
		if err != nil {
			return domain.RunnerConfig{}, rerrors.Wrap(err, "error resolving runner dind")
		}
	}

	runnerProvider, containerRuntime, err := s.resolveRunnerRuntime(ctx, svc.Env, runner.Provider)
	if err != nil {
		return domain.RunnerConfig{}, rerrors.Wrap(err, "error resolving runner runtime")
	}

	err = ensureRunnerRunning(ctx, containerRuntime, name)
	if err != nil {
		return domain.RunnerConfig{}, rerrors.Wrap(err, "error checking runner container")
	}

	config.Settings, err = runnerProvider.ReadSettings(ctx, containerRuntime, name)
	if err != nil {
		return domain.RunnerConfig{}, rerrors.Wrap(err, "error reading runner settings")
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

	err := validateDockerSourceUpdate(req)
	if err != nil {
		return domain.UpdateRunnerConfigResult{}, rerrors.Wrap(err)
	}

	err = validateSettings(req)
	if err != nil {
		return domain.UpdateRunnerConfigResult{}, rerrors.Wrap(err, "error validating runner settings")
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
		DindServiceId:       runner.DindServiceId,
		GitlabRunnerId:      runner.GitlabRunnerId,
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

	if req.DindName.Valid {
		result.RequiresRedeploy = true
	}

	targetDindId, err := s.targetDindServiceId(ctx, runner, req)
	if err != nil {
		return domain.UpdateRunnerConfigResult{}, rerrors.Wrap(err, "error resolving target dind")
	}

	err = s.releaseBuildkit(ctx, svc.Env, req.Name, runner, targetDindId)
	if err != nil {
		return domain.UpdateRunnerConfigResult{}, rerrors.Wrap(err, "error releasing runner buildkit")
	}

	upsertReq.DindServiceId = targetDindId

	if targetDindId != 0 {
		upsertReq.DockerSocketAddress = ""
	}

	if req.Concurrent.Valid {
		err = s.applyConcurrent(ctx, svc.Env, runner.Provider, req.Name, req.Concurrent.Value)
		if err != nil {
			return domain.UpdateRunnerConfigResult{}, rerrors.Wrap(err, "error applying concurrent")
		}

		upsertReq.Concurrent = req.Concurrent.Value
	}

	if hasSettings(req) {
		err = s.applySettings(ctx, svc.Env, runner.Provider, req)
		if err != nil {
			return domain.UpdateRunnerConfigResult{}, rerrors.Wrap(err, "error applying runner settings")
		}
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
	runnerProvider, containerRuntime, err := s.resolveRunnerRuntime(ctx, environment, provider)
	if err != nil {
		return rerrors.Wrap(err, "error resolving runner runtime")
	}

	err = runnerProvider.ApplyConcurrent(ctx, containerRuntime, name, concurrent)
	if err != nil {
		return rerrors.Wrap(err, "error applying provider concurrent")
	}

	return nil
}

// validateDockerSourceUpdate rejects a request that sets both a DinD and a
// tcp:// socket address; unlike create, either one may be absent.
func validateDockerSourceUpdate(req domain.UpdateRunnerConfigReq) error {
	if req.DindName.Value != "" && req.DockerSocketAddress.Value != "" {
		return user_errors.ErrRunnerDockerSourceAmbiguous
	}

	return nil
}

// targetDindServiceId is the DinD the runner row points at once req is
// applied: a named DinD attaches or switches, an empty dind_name detaches, and
// a non-empty docker_socket_address detaches too - the two stay exclusive.
func (s *RunneraasService) targetDindServiceId(
	ctx context.Context, runner domain.Runner, req domain.UpdateRunnerConfigReq,
) (int64, error) {
	if req.DindName.Valid && req.DindName.Value != "" {
		return s.findDindServiceId(ctx, req.DindName.Value)
	}

	isDetach := req.DindName.Valid || (req.DockerSocketAddress.Valid && req.DockerSocketAddress.Value != "")
	if isDetach {
		return 0, nil
	}

	return runner.DindServiceId, nil
}

// releaseBuildkit takes BuildKit off the DinD a runner leaves, while the runner
// row still names it. A switch leaves BuildKit enabled - RedeployRunner deploys
// it on the new DinD once the runner is recreated there; a detach turns it off,
// since BuildKit needs a DinD.
func (s *RunneraasService) releaseBuildkit(
	ctx context.Context, environment, name string, runner domain.Runner, targetDindServiceId int64,
) error {
	isLeavingDind := runner.IsBuildkitEnabled &&
		runner.DindServiceId != 0 &&
		runner.DindServiceId != targetDindServiceId
	if !isLeavingDind {
		return nil
	}

	oldDindName, err := s.dindName(ctx, runner.DindServiceId)
	if err != nil && !rerrors.Is(err, user_errors.ErrDindNotFound) {
		return rerrors.Wrap(err, "error resolving runner dind")
	}

	releaseReq := jobs.ReleaseRunnerBuildkitReq{
		RunnerName:  name,
		Environment: environment,
		Provider:    velez_api.RunnerProvider(velez_api.RunnerProvider_value[runner.Provider]),
		DindName:    oldDindName,
		IsDetach:    targetDindServiceId == 0,
	}

	err = jobs.ReleaseRunnerBuildkit(ctx, s.dataStorage, s.runtimes, releaseReq)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return nil
}
