package jobs

import (
	"context"
	"strings"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/service/secrets"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	DropPgInstanceAction = "drop_pg_instance"

	stepDeletePgInstanceRow     = "delete_pg_instance_row"
	stepRemovePgInstanceService = "remove_pg_instance_service"
	stepDeletePgInstanceSecret  = "delete_pg_instance_secret"
)

type dropPgInstanceHandler struct {
	dataStorage  storage.Storage
	secretsStore secrets.Store
	vervServices service.VervServicesService
}

func NewDropPgInstanceHandler(
	dataStorage storage.Storage,
	secretsStore secrets.Store,
	vervServices service.VervServicesService,
) TaskHandler {
	return &dropPgInstanceHandler{
		dataStorage:  dataStorage,
		secretsStore: secretsStore,
		vervServices: vervServices,
	}
}

func (h *dropPgInstanceHandler) Action() string {
	return DropPgInstanceAction
}

func (h *dropPgInstanceHandler) NewContext() TaskContext {
	return &velez_api.DropPgInstanceTaskPayload{}
}

// The row goes first because it is addressed by the service id, which is
// unresolvable once the service is removed; the secret goes last and is
// addressed by name alone, so no step depends on data an earlier step deleted.
func (h *dropPgInstanceHandler) BuildJobs(taskCtx TaskContext) []NamedJob {
	payload, ok := taskCtx.(*velez_api.DropPgInstanceTaskPayload)
	if !ok {
		panic("drop_pg_instance: BuildJobs called with mismatched TaskContext type")
	}

	name := payload.GetName()
	bareName := strings.TrimPrefix(name, labels.PgaasNamePrefix)

	return []NamedJob{
		{
			Name: stepDeletePgInstanceRow,
			Job: &deletePgInstanceRowJob{
				services:    h.dataStorage.Services(),
				pgInstances: h.dataStorage.PgInstances(),
				name:        name,
			},
		},
		{
			Name: stepRemovePgInstanceService,
			Job: &removePgInstanceServiceJob{
				vervServices: h.vervServices,
				name:         name,
			},
		},
		{
			Name: stepDeletePgInstanceSecret,
			Job: &deletePgInstanceSecretJob{
				secrets:   h.secretsStore,
				secretRef: pgSecretRef(bareName),
			},
		},
	}
}

type deletePgInstanceRowJob struct {
	services    storage.ServicesStorage
	pgInstances storage.PgInstancesStorage
	name        string
}

func (j *deletePgInstanceRowJob) Do(ctx context.Context) error {
	svc, err := j.services.GetByName(ctx, j.name)
	if err != nil {
		if rerrors.Is(err, user_errors.ErrStorageNotFound) {
			return nil
		}

		return rerrors.Wrap(err, "error getting pg instance service")
	}

	err = j.pgInstances.DeletePgInstance(ctx, svc.ID)
	if err != nil && !rerrors.Is(err, user_errors.ErrStorageNotFound) {
		return rerrors.Wrap(err, "error deleting pg instance row")
	}

	return nil
}

type removePgInstanceServiceJob struct {
	vervServices service.VervServicesService
	name         string
}

func (j *removePgInstanceServiceJob) Do(ctx context.Context) error {
	removeReq := domain.RemoveServiceReq{Name: j.name, DropRunningInstances: true}

	err := j.vervServices.Remove(ctx, removeReq)
	if err != nil {
		return rerrors.Wrap(err, "error removing pg instance service")
	}

	return nil
}

type deletePgInstanceSecretJob struct {
	secrets   secrets.Store
	secretRef domain.SecretRef
}

func (j *deletePgInstanceSecretJob) Do(ctx context.Context) error {
	err := j.secrets.Delete(ctx, j.secretRef)
	if err != nil && !rerrors.Is(err, user_errors.ErrSecretNotFound) {
		return rerrors.Wrap(err, "error deleting pg instance secret")
	}

	return nil
}
