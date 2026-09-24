package jobs

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service/secrets"
	"go.vervstack.ru/Velez/internal/service/service_manager/runneraas/providers"
	"go.vervstack.ru/Velez/internal/storage"
)

const (
	ReregisterRunnerAction = "reregister_runner"
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
				req:         payload,
			},
		},
	}
}

// reregisterRunnerJob execs Unregister then Register against the runner's
// already-deployed container, reusing its stored registration token - no
// new container, no new token. Container name == instance/service name, the
// same convention registerRunnerJob (create_runner.go) relies on.
type reregisterRunnerJob struct {
	dataStorage storage.Storage
	secrets     secrets.Store
	runtimes    container_runtime.RuntimeResolver

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

	err = runnerProvider.Unregister(ctx, containerRuntime, name)
	if err != nil {
		return rerrors.Wrap(err, "error unregistering runner")
	}

	err = runnerProvider.Register(ctx, containerRuntime, name, runner.BaseUrl, token, "", name)
	if err != nil {
		return rerrors.Wrap(err, "error registering runner")
	}

	return nil
}
