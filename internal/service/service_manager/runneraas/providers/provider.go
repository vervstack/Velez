// Package providers selects the runneraas Provider implementation for a
// RunnerProvider enum value. Split out from the runneraas package itself so
// both runneraas.RunneraasService (the RPC-facing side, still needed for
// GetRunnerCredentials) and internal/jobs/create_runner.go (which
// RunneraasService.CreateRunner enqueues into, and which does the actual
// token minting/deploy) can resolve a Provider without an import cycle
// between runneraas and internal/jobs.
package providers

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service/service_manager/runneraas/providers/github"
	"go.vervstack.ru/Velez/internal/service/service_manager/runneraas/providers/gitlab"
	"go.vervstack.ru/Velez/internal/user_errors"
)

// Provider mints a short-lived runner registration token from a caller-
// supplied access token and maps registration facts onto a runner
// container's env - the one seam that differs between CI systems. Everything
// else in runneraas/create_runner.go is provider-agnostic; neither ever
// branches on provider outside of resolving one via For.
type Provider interface {
	MintRegistrationToken(
		ctx context.Context, scope velez_api.RunnerScope, target, baseUrl, accessToken string,
	) (string, error)
	RegistrationEnv(
		scope velez_api.RunnerScope, target, baseUrl, runnerName, registrationToken string, labels []string,
	) map[string]string
	// Register finishes registering a deployed runner container, run once
	// waitForRunnerDeployJob confirms it exists. A no-op for a provider that
	// already self-registers from RegistrationEnv (GitHub); for one that
	// doesn't (GitLab), execs its registration command inside containerID via
	// runtime and fails on a non-zero exit code. dockerImage is provider-
	// specific and ignored by a provider that doesn't use it. concurrent is
	// applied through ApplyConcurrent once registration succeeds. Idempotent:
	// a previous registration in the container is unregistered first, so its
	// config always ends with exactly one runner entry.
	Register(
		ctx context.Context, runtime container_runtime.ContainerRuntime,
		containerID, baseUrl, registrationToken, dockerImage, runnerName string, concurrent int32,
	) error
	// ApplyConcurrent writes the provider's global job-concurrency setting
	// into the running container's config; no-op for a provider without one
	// (GitHub).
	ApplyConcurrent(
		ctx context.Context, runtime container_runtime.ContainerRuntime, containerID string, concurrent int32,
	) error
	// ReadSettings reads the provider's config-file-backed settings out of the running container;
	// zero value for a provider without any (GitHub).
	ReadSettings(
		ctx context.Context, runtime container_runtime.ContainerRuntime, containerID string,
	) (domain.GitlabRunnerSettings, error)
	// ApplySettings writes settings into the running container's config, replacing all of them;
	// no-op for a provider without any.
	ApplySettings(
		ctx context.Context, runtime container_runtime.ContainerRuntime, containerID string,
		settings domain.GitlabRunnerSettings,
	) error
	// Unregister undoes a prior Register - execs the provider's unregister
	// command inside containerID so a subsequent Register doesn't append a
	// duplicate local runner entry to the container's config. A no-op for a
	// provider that self-registers (GitHub).
	Unregister(ctx context.Context, runtime container_runtime.ContainerRuntime, containerID string) error
	DescriptorName() string
	DataPath() string
}

// For looks up the Provider for p, wrapped so every caller gets the same
// rerrors-wrapped sentinel for an unrecognized/unspecified RunnerProvider.
func For(p velez_api.RunnerProvider) (Provider, error) {
	switch p {
	case velez_api.RunnerProvider_GITHUB:
		return github.New(), nil
	case velez_api.RunnerProvider_GITLAB:
		return gitlab.New(), nil
	default:
		return nil, rerrors.Wrap(user_errors.ErrRunnerProviderUnsupported)
	}
}
