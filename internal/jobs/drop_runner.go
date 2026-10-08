package jobs

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/service/secrets"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	DropRunnerAction = "drop_runner"

	stepDeleteRunnerRow     = "delete_runner_row"
	stepRemoveRunnerService = "remove_runner_service"
	stepDeleteRunnerSecret  = "delete_runner_secret"
)

type dropRunnerHandler struct {
	dataStorage  storage.Storage
	secretsStore secrets.Store
	vervServices service.VervServicesService
}

func NewDropRunnerHandler(
	dataStorage storage.Storage,
	secretsStore secrets.Store,
	vervServices service.VervServicesService,
) TaskHandler {
	return &dropRunnerHandler{
		dataStorage:  dataStorage,
		secretsStore: secretsStore,
		vervServices: vervServices,
	}
}

func (h *dropRunnerHandler) Action() string {
	return DropRunnerAction
}

func (h *dropRunnerHandler) NewContext() TaskContext {
	return &velez_api.DropRunnerTaskPayload{}
}

// The row goes first because it is addressed by the service id, which is
// unresolvable once the service is removed; the secret goes last and is
// addressed by name alone, so no step depends on data an earlier step deleted.
func (h *dropRunnerHandler) BuildJobs(taskCtx TaskContext) []NamedJob {
	payload, ok := taskCtx.(*velez_api.DropRunnerTaskPayload)
	if !ok {
		panic("drop_runner: BuildJobs called with mismatched TaskContext type")
	}

	name := payload.GetName()

	return []NamedJob{
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
	}
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
