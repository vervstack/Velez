package runneraas

import (
	"context"
	"slices"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service/service_manager/runneraas/providers"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	pullPolicyAlways = "always"
)

var (
	supportedPullPolicies = []string{pullPolicyAlways, "if-not-present", "never"}
	supportedLogLevels    = []string{"debug", "info", "warn", "error", "fatal", "panic"}
)

func (s *RunneraasService) resolveRunnerRuntime(
	ctx context.Context, environment, provider string,
) (providers.Provider, container_runtime.ContainerRuntime, error) {
	providerEnum := velez_api.RunnerProvider(velez_api.RunnerProvider_value[provider])

	runnerProvider, err := providers.For(providerEnum)
	if err != nil {
		return nil, nil, rerrors.Wrap(err, "error resolving runner provider")
	}

	containerRuntime, err := s.runtimes.Runtime(ctx, environment)
	if err != nil {
		return nil, nil, rerrors.Wrap(err, "error resolving container runtime")
	}

	return runnerProvider, containerRuntime, nil
}

// ensureRunnerRunning relies on IsContainerRunning's (isRunning, isFound, err)
// results: an absent container and a stopped one are both "not running".
func ensureRunnerRunning(ctx context.Context, runtime container_runtime.ContainerRuntime, name string) error {
	isRunning, isFound, err := runtime.IsContainerRunning(ctx, name)
	if err != nil {
		return rerrors.Wrap(err, "error checking runner container state")
	}

	if !isFound || !isRunning {
		return rerrors.Wrap(user_errors.ErrRunnerNotRunning)
	}

	return nil
}

func hasSettings(req domain.UpdateRunnerConfigReq) bool {
	return req.PullPolicy.Valid ||
		req.AllowedPullPolicies.Valid ||
		req.CheckInterval.Valid ||
		req.LogLevel.Valid ||
		req.ShutdownTimeout.Valid
}

func (s *RunneraasService) applySettings(
	ctx context.Context, environment, provider string, req domain.UpdateRunnerConfigReq,
) error {
	runnerProvider, containerRuntime, err := s.resolveRunnerRuntime(ctx, environment, provider)
	if err != nil {
		return rerrors.Wrap(err, "error resolving runner runtime")
	}

	err = ensureRunnerRunning(ctx, containerRuntime, req.Name)
	if err != nil {
		return rerrors.Wrap(err, "error checking runner container")
	}

	current, err := runnerProvider.ReadSettings(ctx, containerRuntime, req.Name)
	if err != nil {
		return rerrors.Wrap(err, "error reading runner settings")
	}

	merged := mergeSettings(current, req)

	err = runnerProvider.ApplySettings(ctx, containerRuntime, req.Name, merged)
	if err != nil {
		return rerrors.Wrap(err, "error applying provider settings")
	}

	return nil
}

func mergeSettings(current domain.GitlabRunnerSettings, req domain.UpdateRunnerConfigReq) domain.GitlabRunnerSettings {
	merged := current

	if req.PullPolicy.Valid {
		merged.PullPolicy = req.PullPolicy.Value
	}

	if req.AllowedPullPolicies.Valid {
		merged.AllowedPullPolicies = req.AllowedPullPolicies.Value
	}

	if req.CheckInterval.Valid {
		merged.CheckInterval = req.CheckInterval.Value
	}

	if req.LogLevel.Valid {
		merged.LogLevel = req.LogLevel.Value
	}

	if req.ShutdownTimeout.Valid {
		merged.ShutdownTimeout = req.ShutdownTimeout.Value
	}

	return merged
}

func validateSettings(req domain.UpdateRunnerConfigReq) error {
	if req.PullPolicy.Valid && !isValidPolicyList(req.PullPolicy.Value) {
		return rerrors.Wrap(user_errors.ErrRunnerSettingsInvalid)
	}

	if req.AllowedPullPolicies.Valid && !isValidPolicyList(req.AllowedPullPolicies.Value) {
		return rerrors.Wrap(user_errors.ErrRunnerSettingsInvalid)
	}

	if req.LogLevel.Valid && req.LogLevel.Value != "" && !slices.Contains(supportedLogLevels, req.LogLevel.Value) {
		return rerrors.Wrap(user_errors.ErrRunnerSettingsInvalid)
	}

	if req.CheckInterval.Valid && req.CheckInterval.Value < 0 {
		return rerrors.Wrap(user_errors.ErrRunnerSettingsInvalid)
	}

	if req.ShutdownTimeout.Valid && req.ShutdownTimeout.Value < 0 {
		return rerrors.Wrap(user_errors.ErrRunnerSettingsInvalid)
	}

	return nil
}

func isValidPolicyList(policies []string) bool {
	seen := make(map[string]struct{}, len(policies))

	for _, policy := range policies {
		if !slices.Contains(supportedPullPolicies, policy) {
			return false
		}

		_, isDuplicate := seen[policy]
		if isDuplicate {
			return false
		}

		seen[policy] = struct{}{}
	}

	return true
}
