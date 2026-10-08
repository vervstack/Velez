package jobs

import (
	"context"
	"io/fs"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/gitlab_runner_config"
	"go.vervstack.ru/Velez/internal/service/secrets"
	"go.vervstack.ru/Velez/internal/service/service_manager/runneraas/providers"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	ReregisterRunnerAction = "reregister_runner"

	runnerConfigFileMode fs.FileMode = 0o600
)

// reregisterRunnerRequestAccessor is the narrow TaskContext slice
// reregisterRunnerJob needs to read the original request.
// *velez_api.ReregisterRunnerTaskPayload satisfies it.
type reregisterRunnerRequestAccessor interface {
	GetRequest() *velez_api.ReregisterRunner_Request
}

type reregisterRunnerHandler struct {
	dataStorage  storage.Storage
	secretsStore secrets.Store
	runtimes     container_runtime.RuntimeResolver
}

func NewReregisterRunnerHandler(
	dataStorage storage.Storage,
	secretsStore secrets.Store,
	runtimes container_runtime.RuntimeResolver,
) TaskHandler {
	return &reregisterRunnerHandler{
		dataStorage:  dataStorage,
		secretsStore: secretsStore,
		runtimes:     runtimes,
	}
}

func (h *reregisterRunnerHandler) Action() string {
	return ReregisterRunnerAction
}

func (h *reregisterRunnerHandler) NewContext() TaskContext {
	return &velez_api.ReregisterRunnerTaskPayload{}
}

// BuildJobs is a single job, unlike create_runner's multi-step chain - the
// container already exists and the registration token is already stored, so
// there's nothing to deploy or mint.
func (h *reregisterRunnerHandler) BuildJobs(taskCtx TaskContext) []NamedJob {
	payload, ok := taskCtx.(*velez_api.ReregisterRunnerTaskPayload)
	if !ok {
		panic("reregister_runner: BuildJobs called with mismatched TaskContext type")
	}

	return []NamedJob{
		{
			Name: ReregisterRunnerAction,
			Job: &reregisterRunnerJob{
				dataStorage: h.dataStorage,
				secrets:     h.secretsStore,
				runtimes:    h.runtimes,
				seeder:      configSeederFor(velez_api.RunnerProvider_GITLAB),
				req:         payload,
			},
		},
	}
}

// reregisterRunnerJob re-registers the runner's already-deployed container.
// A GitLab runner with a recorded gitlab id rotates its authentication token
// through the GitLab API and rewrites it into config.toml in a single
// read-modify-write, leaving every other setting untouched. Any other runner
// execs Register (which unregisters the previous entry itself, so config.toml
// ends with exactly one), reusing its stored registration token - no new
// container, no new token. Container name == instance/service name.
type reregisterRunnerJob struct {
	dataStorage storage.Storage
	secrets     secrets.Store
	runtimes    container_runtime.RuntimeResolver
	seeder      providers.ConfigSeeder

	req reregisterRunnerRequestAccessor
}

func (j *reregisterRunnerJob) Do(ctx context.Context) error {
	name := j.req.GetRequest().GetName()

	svc, err := j.dataStorage.Services().GetByName(ctx, name)
	if err != nil {
		return rerrors.Wrap(err, "error getting runner service")
	}

	runner, err := j.dataStorage.Runners().GetRunnerByServiceID(ctx, svc.ID)
	if err != nil {
		return rerrors.Wrap(err, "error getting runner row")
	}

	providerEnum := velez_api.RunnerProvider(velez_api.RunnerProvider_value[runner.Provider])

	if providerEnum == velez_api.RunnerProvider_GITLAB && runner.GitlabRunnerId > 0 {
		return j.rotateGitlabToken(ctx, svc.Env, name, runner.BaseUrl)
	}

	runnerProvider, err := providers.For(providerEnum)
	if err != nil {
		return rerrors.Wrap(err, "error resolving runner provider")
	}

	token, err := j.secrets.Get(ctx, domain.RunnerRegistrationTokenSecretRef(name))
	if err != nil {
		return rerrors.Wrap(err, "error resolving runner registration token")
	}

	containerRuntime, err := j.runtimes.Runtime(ctx, svc.Env)
	if err != nil {
		return rerrors.Wrap(err, "error resolving container runtime")
	}

	// Register replaces the whole [[runners]] entry, including its docker
	// settings, so they are read before it and written back after.
	settings, err := runnerProvider.ReadSettings(ctx, containerRuntime, name)
	if err != nil {
		return rerrors.Wrap(err, "error reading runner settings")
	}

	err = runnerProvider.Register(
		ctx, containerRuntime, name, runner.BaseUrl, token, runner.DockerImage, name,
		effectiveConcurrent(runner.Concurrent),
	)
	if err != nil {
		return rerrors.Wrap(err, "error registering runner")
	}

	err = runnerProvider.ApplySettings(ctx, containerRuntime, name, settings)
	if err != nil {
		return rerrors.Wrap(err, "error restoring runner settings")
	}

	return nil
}

func (j *reregisterRunnerJob) rotateGitlabToken(ctx context.Context, environment, name, baseUrl string) error {
	if j.seeder == nil {
		return rerrors.Wrap(user_errors.ErrRunnerProviderUnsupported)
	}

	containerRuntime, err := j.runtimes.Runtime(ctx, environment)
	if err != nil {
		return rerrors.Wrap(err, "error resolving container runtime")
	}

	token, err := currentGitlabRunnerToken(ctx, containerRuntime, j.secrets, name)
	if err != nil {
		return rerrors.Wrap(err)
	}

	newToken, err := j.seeder.ResetRunnerToken(ctx, baseUrl, token)
	if err != nil {
		return rerrors.Wrap(err, "error resetting gitlab runner token")
	}

	config, err := containerRuntime.CopyFromContainer(ctx, name, gitlab_runner_config.ConfigPath)
	if err != nil {
		return rerrors.Wrap(err, "error reading runner config")
	}

	updated, err := gitlab_runner_config.SetRunnerToken(config, newToken)
	if err != nil {
		return rerrors.Wrap(err, "error setting runner token in config")
	}

	err = containerRuntime.CopyToContainer(ctx, name, gitlab_runner_config.ConfigPath, updated, runnerConfigFileMode)
	if err != nil {
		return rerrors.Wrap(err, "error writing runner config")
	}

	err = j.secrets.Put(ctx, domain.RunnerRegistrationTokenSecretRef(name), newToken)
	if err != nil {
		return rerrors.Wrap(err, "error storing runner registration token")
	}

	return nil
}

// currentGitlabRunnerToken reads the runner's live authentication token from
// config.toml - the source of truth once the runner has rotated it - and falls
// back to the stored secret when the container or file is unavailable.
func currentGitlabRunnerToken(
	ctx context.Context, containerRuntime container_runtime.ContainerRuntime, store secrets.Store, name string,
) (string, error) {
	config, err := containerRuntime.CopyFromContainer(ctx, name, gitlab_runner_config.ConfigPath)
	if err == nil {
		token, isFound := gitlab_runner_config.RunnerToken(config)
		if isFound {
			return token, nil
		}
	}

	token, err := store.Get(ctx, domain.RunnerRegistrationTokenSecretRef(name))
	if err != nil {
		return "", rerrors.Wrap(err, "error resolving runner registration token")
	}

	return token, nil
}
