// Package gitlab implements runneraas.Provider for GitLab. The official
// gitlab/gitlab-runner image doesn't self-register from env vars alone (it
// needs `gitlab-runner register` run once to write config.toml before
// `gitlab-runner run` serves jobs), so this provider deploys a bare
// container and trusts the caller's supplied access token as an
// already-valid registration token, then execs the register command inside
// the container itself once it's deployed (see Register). GitLab's own API
// for minting a registration token (the legacy shared-token reset endpoint)
// is deprecated since 16.0 and disabled by default on 17.0+/gitlab.com, so
// this provider does not call it.
package gitlab

import (
	"bytes"
	"context"
	"errors"
	"slices"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/gitlab_runner_config"
	"go.vervstack.ru/Velez/internal/proxyenv"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	defaultBaseUrl = "https://gitlab.com"

	// defaultDockerImage is the docker executor's job image when the
	// request's GitlabConfig.docker_image is empty.
	defaultDockerImage = "alpine:3.24.2"

	// cacheVolumeSuffix names the docker executor's shared cache volume
	// "<runner container name>-cache". One named volume, mounted at /cache
	// by every job slot, replaces the per-runner-entry anonymous /cache
	// volume gitlab-runner registers by default - that one never survives
	// from one job to the next, so CI caches never hit.
	cacheVolumeSuffix = "-cache"
	cacheMountPath    = "/cache"

	pullPolicyIfNotPresent = "if-not-present"
	pullPolicyAlways       = "always"

	// registerExecutor is always "docker" - not caller-configurable, see
	// GitlabConfig.docker_image's doc comment.
	registerExecutor = "docker"

	// gitlabRunnerBin is the official image's CLI binary, execed for both
	// Register and Unregister.
	gitlabRunnerBin = "gitlab-runner"

	registerCommand   = "register"
	unregisterCommand = "unregister"

	descriptorName = "gitlab_runner"

	// dataPath is gitlab-runner's config/registration directory, so a
	// registration written by Register survives a container restart.
	dataPath = gitlab_runner_config.DataPath

	configPath = gitlab_runner_config.ConfigPath

	configFileMode = 0o600

	// jobDockerAlias is the hostname job containers reach the docker:dind
	// service under; it must never go through the proxy.
	jobDockerAlias = "docker"
)

// Provider implements runneraas.Provider for GitLab.
type Provider struct{}

// New builds a Provider.
func New() *Provider {
	return &Provider{}
}

func (p *Provider) DescriptorName() string {
	return descriptorName
}

func (p *Provider) DataPath() string {
	return dataPath
}

// MintRegistrationToken is a pass-through in v1: it trusts the caller's
// accessToken as an already-valid GitLab runner registration token and
// returns it unchanged. A future PR wires GitLab's real token-minting API
// instead of trusting the caller's raw input (see the package doc comment
// for why v1 doesn't call GitLab's API today).
func (p *Provider) MintRegistrationToken(
	_ context.Context, _ velez_api.RunnerScope, _, _, accessToken string,
) (string, error) {
	if accessToken == "" {
		return "", user_errors.ErrGitlabAccessTokenEmpty
	}

	return accessToken, nil
}

// RegistrationEnv returns no env vars - v1 has no automated registration
// step for GitLab (see the package doc comment), so the container is
// deployed bare and every argument here is unused.
func (p *Provider) RegistrationEnv(
	_ velez_api.RunnerScope, _, _, _, _ string, _ []string,
) map[string]string {
	return map[string]string{}
}

// Register execs `gitlab-runner register --non-interactive` inside
// containerID via runtime, finishing what the package doc comment describes
// - config.toml doesn't exist until this runs once. Fails on a non-zero
// exit code; the register command's own output is never folded into the
// returned error, since it can carry the caller's GitLab server's response
// text (untrusted external content).
//
// Register is idempotent: `gitlab-runner register` only ever appends a
// [[runners]] entry, so any entry already in config.toml is cleared first
// (clearRegistrations) and the result is verified to hold exactly one. A
// container's config.toml therefore always ends with a single entry no
// matter how often Register runs against it - a retried create job and
// reregister both funnel through here.
func (p *Provider) Register(
	ctx context.Context, runtime container_runtime.ContainerRuntime,
	containerID, baseUrl, registrationToken, dockerImage, runnerName string, concurrent int32,
) error {
	base := baseUrl
	if base == "" {
		base = defaultBaseUrl
	}

	image := dockerImage
	if image == "" {
		image = defaultDockerImage
	}

	err := p.clearRegistrations(ctx, runtime, containerID)
	if err != nil {
		return rerrors.Wrap(err, "error clearing previous gitlab-runner registrations")
	}

	execOpts := container.ExecOptions{
		Cmd: []string{
			gitlabRunnerBin, registerCommand,
			"--non-interactive",
			"--url", base,
			"--registration-token", registrationToken,
			"--executor", registerExecutor,
			"--docker-image", image,
			"--docker-privileged",
			"--description", runnerName,
		},
		AttachStdout: true,
		AttachStderr: true,
	}

	_, exitCode, err := runtime.Exec(ctx, containerID, execOpts)
	if err != nil {
		return rerrors.Wrap(err, "error executing gitlab-runner register")
	}

	if exitCode != 0 {
		return rerrors.Wrap(user_errors.ErrGitlabRunnerRegisterFailed)
	}

	err = p.applyRegistrationDefaults(ctx, runtime, containerID, concurrent)
	if err != nil {
		return rerrors.Wrap(err, "error applying defaults after gitlab-runner register")
	}

	return nil
}

// ApplyConcurrent rewrites config.toml's global `concurrent` key inside
// containerID. gitlab-runner has no register/run flag for it, and hot-reloads
// config.toml on change.
func (p *Provider) ApplyConcurrent(
	ctx context.Context, runtime container_runtime.ContainerRuntime, containerID string, concurrent int32,
) error {
	config, err := runtime.CopyFromContainer(ctx, containerID, configPath)
	if err != nil {
		return rerrors.Wrap(err, "error reading gitlab-runner config.toml")
	}

	updated := gitlab_runner_config.SetConcurrent(config, concurrent)

	err = runtime.CopyToContainer(ctx, containerID, configPath, updated, configFileMode)
	if err != nil {
		return rerrors.Wrap(err, "error writing gitlab-runner config.toml")
	}

	return nil
}

// ApplyNetworkMode sets the docker network the job containers of containerID
// join in config.toml's `[runners.docker]`; an empty networkMode removes it.
// Job containers are created on the daemon the runner talks to, so the network
// has to exist there. gitlab-runner hot-reloads the file. An unregistered
// container has no config.toml: nothing to clear, an error to set.
func (p *Provider) ApplyNetworkMode(
	ctx context.Context, runtime container_runtime.ContainerRuntime, containerID, networkMode string,
) error {
	config, isFound, err := p.readConfig(ctx, runtime, containerID)
	if err != nil {
		return rerrors.Wrap(err, "error reading gitlab-runner config.toml")
	}

	if !isFound && networkMode == "" {
		return nil
	}

	if !isFound {
		return rerrors.Wrap(user_errors.ErrContainerFileNotFound, configPath)
	}

	updated, err := gitlab_runner_config.SetNetworkMode(config, networkMode)
	if err != nil {
		return rerrors.Wrap(err, "error applying network mode")
	}

	if bytes.Equal(updated, config) {
		return nil
	}

	err = runtime.CopyToContainer(ctx, containerID, configPath, updated, configFileMode)
	if err != nil {
		return rerrors.Wrap(err, "error writing gitlab-runner config.toml")
	}

	return nil
}

// ReadSettings parses the pull-policy, check-interval, log-level and
// shutdown-timeout keys out of containerID's config.toml.
func (p *Provider) ReadSettings(
	ctx context.Context, runtime container_runtime.ContainerRuntime, containerID string,
) (domain.GitlabRunnerSettings, error) {
	config, err := runtime.CopyFromContainer(ctx, containerID, configPath)
	if err != nil {
		return domain.GitlabRunnerSettings{}, rerrors.Wrap(err, "error reading gitlab-runner config.toml")
	}

	settings, err := gitlab_runner_config.ReadSettings(config)
	if err != nil {
		return domain.GitlabRunnerSettings{}, rerrors.Wrap(err, "error parsing gitlab-runner config.toml")
	}

	return settings, nil
}

// ApplySettings replaces every setting ReadSettings reports in containerID's
// config.toml; gitlab-runner hot-reloads the file on change.
func (p *Provider) ApplySettings(
	ctx context.Context, runtime container_runtime.ContainerRuntime, containerID string,
	settings domain.GitlabRunnerSettings,
) error {
	config, err := runtime.CopyFromContainer(ctx, containerID, configPath)
	if err != nil {
		return rerrors.Wrap(err, "error reading gitlab-runner config.toml")
	}

	updated, err := gitlab_runner_config.ApplySettings(config, settings)
	if err != nil {
		return rerrors.Wrap(err, "error applying gitlab-runner settings")
	}

	err = runtime.CopyToContainer(ctx, containerID, configPath, updated, configFileMode)
	if err != nil {
		return rerrors.Wrap(err, "error writing gitlab-runner config.toml")
	}

	return nil
}

// SyncProxy writes the proxy env of containerID into config.toml's
// `[[runners]] environment`, which is the env gitlab-runner gives every job
// container - they are spawned by the runner, not by Velez, so they never see
// the runner container's own env. The container env stays the source of truth:
// no proxy there removes the entries again. gitlab-runner hot-reloads the
// file. A container that has not registered yet has no config.toml and is left
// alone - Register applies the proxy itself.
func (p *Provider) SyncProxy(
	ctx context.Context, runtime container_runtime.ContainerRuntime, containerID string,
) error {
	config, isFound, err := p.readConfig(ctx, runtime, containerID)
	if err != nil {
		return rerrors.Wrap(err, "error reading gitlab-runner config.toml")
	}

	if !isFound {
		return nil
	}

	updated, err := p.withJobProxy(ctx, runtime, containerID, config)
	if err != nil {
		return rerrors.Wrap(err, "error applying job proxy")
	}

	if bytes.Equal(updated, config) {
		return nil
	}

	err = runtime.CopyToContainer(ctx, containerID, configPath, updated, configFileMode)
	if err != nil {
		return rerrors.Wrap(err, "error writing gitlab-runner config.toml")
	}

	return nil
}

// Unregister execs `gitlab-runner unregister --all-runners` inside
// containerID, clearing every [[runners]] entry config.toml currently holds
// so a following Register doesn't append a duplicate local entry (which
// would double-run every CI job the container picks up). Fails on a
// non-zero exit code, same discipline as Register.
func (p *Provider) Unregister(
	ctx context.Context, runtime container_runtime.ContainerRuntime, containerID string,
) error {
	execOpts := container.ExecOptions{
		Cmd:          []string{gitlabRunnerBin, unregisterCommand, "--all-runners"},
		AttachStdout: true,
		AttachStderr: true,
	}

	_, exitCode, err := runtime.Exec(ctx, containerID, execOpts)
	if err != nil {
		return rerrors.Wrap(err, "error executing gitlab-runner unregister")
	}

	if exitCode != 0 {
		return rerrors.Wrap(user_errors.ErrGitlabRunnerUnregisterFailed)
	}

	return nil
}

// clearRegistrations leaves containerID's config.toml without any [[runners]]
// entry. It asks gitlab-runner to unregister them first, so the runner is
// also removed on the GitLab side, then drops whatever entry is still left:
// `unregister --all-runners` keeps the local entry of a runner GitLab refused
// to delete (token already revoked, runner already gone) and still exits 0,
// and a surviving entry is exactly what the next `register` would duplicate.
func (p *Provider) clearRegistrations(
	ctx context.Context, runtime container_runtime.ContainerRuntime, containerID string,
) error {
	config, isFound, err := p.readConfig(ctx, runtime, containerID)
	if err != nil {
		return rerrors.Wrap(err, "error reading gitlab-runner config.toml")
	}

	if !isFound || gitlab_runner_config.CountRunners(config) == 0 {
		return nil
	}

	err = p.Unregister(ctx, runtime, containerID)
	if err != nil {
		return rerrors.Wrap(err, "error unregistering previous runners")
	}

	config, err = runtime.CopyFromContainer(ctx, containerID, configPath)
	if err != nil {
		return rerrors.Wrap(err, "error reading gitlab-runner config.toml after unregister")
	}

	if gitlab_runner_config.CountRunners(config) == 0 {
		return nil
	}

	cleared := gitlab_runner_config.RemoveRunners(config)

	err = runtime.CopyToContainer(ctx, containerID, configPath, cleared, configFileMode)
	if err != nil {
		return rerrors.Wrap(err, "error writing gitlab-runner config.toml")
	}

	return nil
}

// readConfig reads config.toml; isFound is false on a container that has not
// registered yet, where the file does not exist.
func (p *Provider) readConfig(
	ctx context.Context, runtime container_runtime.ContainerRuntime, containerID string,
) (config []byte, isFound bool, err error) {
	config, err = runtime.CopyFromContainer(ctx, containerID, configPath)
	if errors.Is(err, user_errors.ErrContainerFileNotFound) || errdefs.IsNotFound(err) {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, rerrors.Wrap(err)
	}

	return config, true, nil
}

// applyRegistrationDefaults writes everything `gitlab-runner register` has no
// flag for - or whose default it would add to rather than replace (the
// anonymous /cache volume) - into the just-registered config.toml in one
// read-modify-write, and verifies the file holds exactly one [[runners]] entry.
func (p *Provider) applyRegistrationDefaults(
	ctx context.Context, runtime container_runtime.ContainerRuntime, containerID string, concurrent int32,
) error {
	config, err := runtime.CopyFromContainer(ctx, containerID, configPath)
	if err != nil {
		return rerrors.Wrap(err, "error reading gitlab-runner config.toml")
	}

	runnersCount := gitlab_runner_config.CountRunners(config)
	if runnersCount != 1 {
		return rerrors.Wrapf(user_errors.ErrGitlabRunnerEntryCount, "found %d [[runners]] entries", runnersCount)
	}

	defaults := gitlab_runner_config.DockerDefaults{
		Volumes:             []string{containerID + cacheVolumeSuffix + ":" + cacheMountPath},
		PullPolicy:          []string{pullPolicyIfNotPresent},
		AllowedPullPolicies: []string{pullPolicyIfNotPresent, pullPolicyAlways},
	}

	withDefaults, err := gitlab_runner_config.SetDockerDefaults(config, defaults)
	if err != nil {
		return rerrors.Wrap(err, "error applying docker defaults")
	}

	withConcurrent := gitlab_runner_config.SetConcurrent(withDefaults, concurrent)

	updated, err := p.withJobProxy(ctx, runtime, containerID, withConcurrent)
	if err != nil {
		return rerrors.Wrap(err, "error applying job proxy")
	}

	err = runtime.CopyToContainer(ctx, containerID, configPath, updated, configFileMode)
	if err != nil {
		return rerrors.Wrap(err, "error writing gitlab-runner config.toml")
	}

	return nil
}

// withJobProxy returns config with the proxy env of containerID applied to the
// runner entry's job environment.
func (p *Provider) withJobProxy(
	ctx context.Context, runtime container_runtime.ContainerRuntime, containerID string, config []byte,
) ([]byte, error) {
	info, isFound, err := runtime.Inspect(ctx, containerID)
	if err != nil {
		return nil, rerrors.Wrap(err, "error inspecting runner container")
	}

	if !isFound {
		return nil, rerrors.Wrap(user_errors.ErrContainerNotFoundInEnvironment)
	}

	var containerEnv []string

	if info.Config != nil {
		containerEnv = info.Config.Env
	}

	proxyUrl, bypassHosts := proxyenv.ParseList(containerEnv)
	jobBypassHosts := append(slices.Clone(bypassHosts), jobDockerAlias)

	updated, err := gitlab_runner_config.SetRunnerEnvironment(
		config, proxyenv.Keys(), proxyenv.Env(proxyUrl, jobBypassHosts))
	if err != nil {
		return nil, rerrors.Wrap(err, "error setting runner job environment")
	}

	return updated, nil
}
