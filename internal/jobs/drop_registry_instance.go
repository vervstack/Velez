package jobs

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/garage"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/service/secrets"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/registries"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	DropRegistryInstanceAction = "drop_registry_instance"

	stepRemoveRegistryS3Access    = "remove_s3_access"
	stepDeleteRegistryInstanceRow = "delete_registry_instance_row"
	stepRemoveRegistryUiService   = "remove_registry_ui_service"
	stepRemoveRegistryService     = "remove_registry_service"
	stepDeleteBuiltinRegistryRow  = "delete_builtin_registry_row"
	stepDeleteRegistrySecret      = "delete_registry_password_secret"
)

type dropRegistryInstanceHandler struct {
	dataStorage    storage.Storage
	secretsStore   secrets.Store
	vervServices   service.VervServicesService
	runtimes       container_runtime.RuntimeResolver
	configResolver service.ServiceConfigResolver
}

func NewDropRegistryInstanceHandler(
	dataStorage storage.Storage,
	secretsStore secrets.Store,
	vervServices service.VervServicesService,
	runtimes container_runtime.RuntimeResolver,
	configResolver service.ServiceConfigResolver,
) TaskHandler {
	return &dropRegistryInstanceHandler{
		dataStorage:    dataStorage,
		secretsStore:   secretsStore,
		vervServices:   vervServices,
		runtimes:       runtimes,
		configResolver: configResolver,
	}
}

func (h *dropRegistryInstanceHandler) Action() string {
	return DropRegistryInstanceAction
}

func (h *dropRegistryInstanceHandler) NewContext() TaskContext {
	return &velez_api.DropRegistryInstanceTaskPayload{}
}

// BuildJobs orders the steps so each one re-derives what it needs from the
// name alone after a crash: everything that reads the main service (its row id
// for the instance row, its environment for the S3 config and the removals) runs
// before that service is removed, and the main service is removed last among
// the steps that look it up. The builtin registry row and the password secret
// are keyed by name, so they come after the services are gone.
func (h *dropRegistryInstanceHandler) BuildJobs(taskCtx TaskContext) []NamedJob {
	payload, ok := taskCtx.(*velez_api.DropRegistryInstanceTaskPayload)
	if !ok {
		panic("drop_registry_instance: BuildJobs called with mismatched TaskContext type")
	}

	return []NamedJob{
		{
			Name: stepRemoveRegistryS3Access,
			Job: &removeRegistryS3AccessJob{
				services:       h.dataStorage.Services(),
				dependencies:   h.dataStorage.ServiceDependencies(),
				secrets:        h.secretsStore,
				runtimes:       h.runtimes,
				configResolver: h.configResolver,
				req:            payload,
			},
		},
		{
			Name: stepDeleteRegistryInstanceRow,
			Job: &deleteRegistryInstanceRowJob{
				services:          h.dataStorage.Services(),
				registryInstances: h.dataStorage.RegistryInstances(),
				req:               payload,
			},
		},
		{
			Name: stepRemoveRegistryUiService,
			Job: &removeRegistryServiceJob{
				services:     h.dataStorage.Services(),
				vervServices: h.vervServices,
				req:          payload,
				isUiSidecar:  true,
			},
		},
		{
			Name: stepRemoveRegistryService,
			Job: &removeRegistryServiceJob{
				services:     h.dataStorage.Services(),
				vervServices: h.vervServices,
				req:          payload,
			},
		},
		{
			Name: stepDeleteBuiltinRegistryRow,
			Job: &deleteBuiltinRegistryRowJob{
				registries: h.dataStorage.Registries(),
				req:        payload,
			},
		},
		{
			Name: stepDeleteRegistrySecret,
			Job: &deleteRegistryPasswordSecretJob{
				secrets: h.secretsStore,
				req:     payload,
			},
		},
	}
}

func findServiceByName(
	ctx context.Context, services storage.ServicesStorage, name string,
) (domain.Service, bool, error) {
	svc, err := services.GetByName(ctx, name)
	if err != nil {
		if rerrors.Is(err, user_errors.ErrStorageNotFound) {
			return domain.Service{}, false, nil
		}

		return domain.Service{}, false, rerrors.Wrap(err, "error getting service")
	}

	return svc, true, nil
}

type removeRegistryS3AccessJob struct {
	services       storage.ServicesStorage
	dependencies   storage.ServiceDependenciesStorage
	secrets        secrets.Store
	runtimes       container_runtime.RuntimeResolver
	configResolver service.ServiceConfigResolver

	req serviceNameAccessor
}

func (j *removeRegistryS3AccessJob) Do(ctx context.Context) error {
	name := j.req.GetName()

	svc, isFound, err := findServiceByName(ctx, j.services, name)
	if err != nil {
		return rerrors.Wrap(err)
	}

	if !isFound {
		return nil
	}

	env, err := j.configResolver.ReadEnv(ctx, name, svc.Env)
	if err != nil {
		if rerrors.Is(err, user_errors.ErrNoSuchContainer) {
			return nil
		}

		return rerrors.Wrap(err, "error reading registry config")
	}

	binding, isS3 := domain.RegistryS3BindingFromEnv(env)
	if !isS3 {
		return nil
	}

	client, err := newGarageClient(ctx, j.runtimes, j.secrets, svc.Env, binding.InstanceName)
	if err != nil {
		return rerrors.Wrap(err, "error creating garage client")
	}

	err = client.DeleteKey(ctx, binding.AccessKeyId)
	if err != nil && !rerrors.Is(err, garage.ErrNotFound) {
		return rerrors.Wrap(err, "error deleting registry s3 key")
	}

	keyRef := domain.S3KeySecretRef(binding.InstanceName, binding.AccessKeyId)

	err = j.secrets.Delete(ctx, keyRef)
	if err != nil && !rerrors.Is(err, user_errors.ErrSecretNotFound) {
		return rerrors.Wrap(err, "error deleting registry s3 key secret")
	}

	err = j.dependencies.DeleteDependenciesOf(ctx, name)
	if err != nil {
		return rerrors.Wrap(err, "error deleting registry dependencies")
	}

	err = j.configResolver.Delete(ctx, name)
	if err != nil {
		return rerrors.Wrap(err, "error deleting registry config")
	}

	return nil
}

type deleteRegistryInstanceRowJob struct {
	services          storage.ServicesStorage
	registryInstances storage.RegistryInstancesStorage

	req serviceNameAccessor
}

func (j *deleteRegistryInstanceRowJob) Do(ctx context.Context) error {
	svc, isFound, err := findServiceByName(ctx, j.services, j.req.GetName())
	if err != nil {
		return rerrors.Wrap(err)
	}

	if !isFound {
		return nil
	}

	err = j.registryInstances.DeleteRegistryInstance(ctx, svc.ID)
	if err != nil {
		return rerrors.Wrap(err, "error deleting registry instance row")
	}

	return nil
}

type removeRegistryServiceJob struct {
	services     storage.ServicesStorage
	vervServices service.VervServicesService

	req         serviceNameAccessor
	isUiSidecar bool
}

func (j *removeRegistryServiceJob) Do(ctx context.Context) error {
	name := j.req.GetName()

	svc, isFound, err := findServiceByName(ctx, j.services, name)
	if err != nil {
		return rerrors.Wrap(err)
	}

	if !isFound {
		return nil
	}

	targetName := name
	if j.isUiSidecar {
		targetName = registryaasUiServiceName(name)
	}

	removeReq := domain.RemoveServiceReq{
		Name:                 targetName,
		DropRunningInstances: true,
		Environment:          svc.Env,
	}

	err = j.vervServices.Remove(ctx, removeReq)
	if err != nil {
		return rerrors.Wrapf(err, "error removing service %s", targetName)
	}

	return nil
}

type deleteBuiltinRegistryRowJob struct {
	registries storage.RegistriesStorage

	req serviceNameAccessor
}

func (j *deleteBuiltinRegistryRowJob) Do(ctx context.Context) error {
	deleter, ok := j.registries.(registries.BuiltinRegistryDeleter)
	if !ok {
		return rerrors.Wrap(user_errors.ErrRegistriesStorageMissingBuiltinDelete)
	}

	err := deleter.DeleteBuiltinRegistry(ctx, j.req.GetName())
	if err != nil {
		return rerrors.Wrap(err, "error deleting registry row")
	}

	return nil
}

type deleteRegistryPasswordSecretJob struct {
	secrets secrets.Store

	req serviceNameAccessor
}

func (j *deleteRegistryPasswordSecretJob) Do(ctx context.Context) error {
	secretRef := registryInstanceSecretRef(j.req.GetName())

	err := j.secrets.Delete(ctx, secretRef)
	if err != nil && !rerrors.Is(err, user_errors.ErrSecretNotFound) {
		return rerrors.Wrap(err, "error deleting registry instance secret")
	}

	return nil
}
