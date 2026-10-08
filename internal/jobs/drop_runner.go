package jobs

import (
	"context"

	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/service/secrets"
	"go.vervstack.ru/Velez/internal/service/service_manager/runneraas/providers"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	DropRunnerAction = "drop_runner"

	stepUnregisterRunner    = "unregister_runner"
	stepDropRunnerBuildkit  = "drop_runner_buildkit"
	stepDeleteRunnerRow     = "delete_runner_row"
	stepRemoveRunnerService = "remove_runner_service"
	stepDeleteRunnerSecret  = "delete_runner_secret"

	stepDeleteRunnerRegistrationToken = "delete_runner_registration_token"
)

type dropRunnerHandler struct {
	dataStorage  storage.Storage
	secretsStore secrets.Store
	vervServices service.VervServicesService
	runtimes     container_runtime.RuntimeResolver
}

func NewDropRunnerHandler(
	dataStorage storage.Storage,
	secretsStore secrets.Store,
	vervServices service.VervServicesService,
	runtimes container_runtime.RuntimeResolver,
) TaskHandler {
	return &dropRunnerHandler{
		dataStorage:  dataStorage,
		secretsStore: secretsStore,
		vervServices: vervServices,
		runtimes:     runtimes,
	}
}

func (h *dropRunnerHandler) Action() string {
	return DropRunnerAction
}

func (h *dropRunnerHandler) NewContext() TaskContext {
	return &velez_api.DropRunnerTaskPayload{}
}

// Unregistering and the BuildKit teardown go before the row because they read
// the runner's gitlab id, base url and BuildKit flag from it; the row goes next because it is addressed by the
// service id, which is unresolvable once the service is removed; the secret
// goes last and is addressed by name alone, so no step depends on data an
// earlier step deleted.
func (h *dropRunnerHandler) BuildJobs(taskCtx TaskContext) []NamedJob {
	payload, ok := taskCtx.(*velez_api.DropRunnerTaskPayload)
	if !ok {
		panic("drop_runner: BuildJobs called with mismatched TaskContext type")
	}

	name := payload.GetName()

	return []NamedJob{
		{
			Name: stepUnregisterRunner,
			Job: &unregisterRunnerJob{
				services: h.dataStorage.Services(),
				runners:  h.dataStorage.Runners(),
				secrets:  h.secretsStore,
				runtimes: h.runtimes,
				seeder:   configSeederFor(velez_api.RunnerProvider_GITLAB),
				name:     name,
			},
		},
		{
			Name: stepDropRunnerBuildkit,
			Job: &dropRunnerBuildkitJob{
				services: h.dataStorage.Services(),
				runners:  h.dataStorage.Runners(),
				runtimes: h.runtimes,
				name:     name,
			},
		},
		{
			Name: stepDeleteRunnerRow,
			Job: &deleteRunnerRowJob{
				services: h.dataStorage.Services(),
				runners:  h.dataStorage.Runners(),
				name:     name,
			},
		},
		{
			Name: stepRemoveRunnerService,
			Job: &removeRunnerServiceJob{
				vervServices: h.vervServices,
				name:         name,
			},
		},
		{
			Name: stepDeleteRunnerSecret,
			Job: &deleteRunnerSecretJob{
				secrets:   h.secretsStore,
				secretRef: domain.RunnerAccessTokenSecretRef(name),
			},
		},
		{
			Name: stepDeleteRunnerRegistrationToken,
			Job: &deleteRunnerSecretJob{
				secrets:   h.secretsStore,
				secretRef: domain.RunnerRegistrationTokenSecretRef(name),
			},
		},
	}
}

// unregisterRunnerJob deletes the runner from GitLab so it does not linger
// there after its container is gone. Best-effort: a failure is logged and
// never blocks the drop. Runners without a recorded gitlab id are skipped.
type unregisterRunnerJob struct {
	services storage.ServicesStorage
	runners  storage.RunnersStorage
	secrets  secrets.Store
	runtimes container_runtime.RuntimeResolver
	seeder   providers.ConfigSeeder
	name     string
}

func (j *unregisterRunnerJob) Do(ctx context.Context) error {
	err := j.unregister(ctx)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Str("runner", j.name).Msg("error unregistering runner in gitlab")
	}

	return nil
}

func (j *unregisterRunnerJob) unregister(ctx context.Context) error {
	svc, err := j.services.GetByName(ctx, j.name)
	if err != nil {
		if rerrors.Is(err, user_errors.ErrStorageNotFound) {
			return nil
		}

		return rerrors.Wrap(err, "error getting runner service")
	}

	runner, err := j.runners.GetRunnerByServiceID(ctx, svc.ID)
	if err != nil {
		if rerrors.Is(err, user_errors.ErrStorageNotFound) {
			return nil
		}

		return rerrors.Wrap(err, "error getting runner row")
	}

	providerEnum := velez_api.RunnerProvider(velez_api.RunnerProvider_value[runner.Provider])
	if providerEnum != velez_api.RunnerProvider_GITLAB || runner.GitlabRunnerId <= 0 {
		return nil
	}

	if j.seeder == nil {
		return rerrors.Wrap(user_errors.ErrRunnerProviderUnsupported)
	}

	containerRuntime, err := j.runtimes.Runtime(ctx, svc.Env)
	if err != nil {
		return rerrors.Wrap(err, "error resolving container runtime")
	}

	token, err := currentGitlabRunnerToken(ctx, containerRuntime, j.secrets, j.name)
	if err != nil {
		return rerrors.Wrap(err)
	}

	err = j.seeder.DeleteRunnerByToken(ctx, runner.BaseUrl, token)
	if err != nil {
		return rerrors.Wrap(err, "error deleting gitlab runner")
	}

	return nil
}

type deleteRunnerRowJob struct {
	services storage.ServicesStorage
	runners  storage.RunnersStorage
	name     string
}

func (j *deleteRunnerRowJob) Do(ctx context.Context) error {
	svc, err := j.services.GetByName(ctx, j.name)
	if err != nil {
		if rerrors.Is(err, user_errors.ErrStorageNotFound) {
			return nil
		}

		return rerrors.Wrap(err, "error getting runner service")
	}

	err = j.runners.DeleteRunner(ctx, svc.ID)
	if err != nil && !rerrors.Is(err, user_errors.ErrStorageNotFound) {
		return rerrors.Wrap(err, "error deleting runner row")
	}

	return nil
}

type removeRunnerServiceJob struct {
	vervServices service.VervServicesService
	name         string
}

func (j *removeRunnerServiceJob) Do(ctx context.Context) error {
	removeReq := domain.RemoveServiceReq{Name: j.name, DropRunningInstances: true}

	err := j.vervServices.Remove(ctx, removeReq)
	if err != nil {
		return rerrors.Wrap(err, "error removing runner service")
	}

	return nil
}

type deleteRunnerSecretJob struct {
	secrets   secrets.Store
	secretRef domain.SecretRef
}

func (j *deleteRunnerSecretJob) Do(ctx context.Context) error {
	err := j.secrets.Delete(ctx, j.secretRef)
	if err != nil && !rerrors.Is(err, user_errors.ErrSecretNotFound) {
		return rerrors.Wrap(err, "error deleting runner secret")
	}

	return nil
}
