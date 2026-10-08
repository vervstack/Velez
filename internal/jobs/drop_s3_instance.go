package jobs

import (
	"context"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/garage"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/service/secrets"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	DropS3InstanceAction = "drop_s3_instance"

	stepUnbindBucketOwners   = "unbind_bucket_owners"
	stepRemoveWebUi          = "remove_web_ui"
	stepRemoveS3Service      = "remove_service"
	stepRemoveDockerResource = "remove_docker_resources"
	stepDeleteS3Secrets      = "delete_secrets"
	stepDeleteConfigEntries  = "delete_config_entries"
)

type dropS3InstanceHandler struct {
	nodeClients    node_clients.NodeClients
	dataStorage    storage.Storage
	secretsStore   secrets.Store
	vervServices   service.VervServicesService
	configResolver service.ServiceConfigResolver
	runtimes       container_runtime.RuntimeResolver
}

func NewDropS3InstanceHandler(
	nodeClients node_clients.NodeClients,
	dataStorage storage.Storage,
	secretsStore secrets.Store,
	vervServices service.VervServicesService,
	configResolver service.ServiceConfigResolver,
	runtimes container_runtime.RuntimeResolver,
) TaskHandler {
	return &dropS3InstanceHandler{
		nodeClients:    nodeClients,
		dataStorage:    dataStorage,
		secretsStore:   secretsStore,
		vervServices:   vervServices,
		configResolver: configResolver,
		runtimes:       runtimes,
	}
}

func (h *dropS3InstanceHandler) Action() string {
	return DropS3InstanceAction
}

func (h *dropS3InstanceHandler) NewContext() TaskContext {
	return &velez_api.DropS3InstanceTaskPayload{}
}

func (h *dropS3InstanceHandler) BuildJobs(taskCtx TaskContext) []NamedJob {
	payload, ok := taskCtx.(*velez_api.DropS3InstanceTaskPayload)
	if !ok {
		panic("drop_s3_instance: BuildJobs called with mismatched TaskContext type")
	}

	locator := &s3InstanceLocator{
		runtimes:    h.runtimes,
		dataStorage: h.dataStorage,
		name:        payload.GetName(),
	}

	return []NamedJob{
		{
			Name: stepUnbindBucketOwners,
			Job: &unbindS3BucketOwnersJob{
				locator:     locator,
				secrets:     h.secretsStore,
				dataStorage: h.dataStorage,
			},
		},
		{
			Name: stepRemoveWebUi,
			Job:  &removeS3WebUiJob{locator: locator, vervServices: h.vervServices},
		},
		{
			Name: stepRemoveS3Service,
			Job:  &removeS3ServiceJob{locator: locator, vervServices: h.vervServices},
		},
		{
			Name: stepRemoveDockerResource,
			Job:  &removeS3DockerResourcesJob{nodeClients: h.nodeClients, name: payload.GetName()},
		},
		{
			Name: stepDeleteS3Secrets,
			Job:  &deleteS3SecretsJob{secrets: h.secretsStore, name: payload.GetName()},
		},
		{
			Name: stepDeleteConfigEntries,
			Job: &deleteS3ConfigEntriesJob{
				configResolver: h.configResolver,
				dataStorage:    h.dataStorage,
				name:           payload.GetName(),
			},
		},
	}
}

type s3InstanceTarget struct {
	environment string
	runtime     container_runtime.ContainerRuntime
	hasGarage   bool
	webUi       *container.Summary
}

func (t s3InstanceTarget) isWebUiSidecar() bool {
	if t.webUi == nil {
		return false
	}

	_, isSidecar := t.webUi.Labels[labels.Sidecar]

	return isSidecar
}

type s3InstanceLocator struct {
	runtimes    container_runtime.RuntimeResolver
	dataStorage storage.Storage
	name        string
}

func (l *s3InstanceLocator) locate(ctx context.Context) (s3InstanceTarget, error) {
	environments, err := l.dataStorage.Environments().ListEnvironments(ctx)
	if err != nil {
		return s3InstanceTarget{}, rerrors.Wrap(err, "error listing environments")
	}

	for _, environment := range environments {
		target, isFound, findErr := l.locateInEnvironment(ctx, environment.Name)
		if findErr != nil {
			return s3InstanceTarget{}, rerrors.Wrap(findErr)
		}

		if isFound {
			return target, nil
		}
	}

	svc, err := l.dataStorage.Services().GetByName(ctx, domain.S3ServiceName(l.name))
	if err != nil {
		if rerrors.Is(err, user_errors.ErrStorageNotFound) {
			return s3InstanceTarget{}, nil
		}

		return s3InstanceTarget{}, rerrors.Wrap(err, "error getting s3 service")
	}

	return s3InstanceTarget{environment: svc.Env}, nil
}

func (l *s3InstanceLocator) locateInEnvironment(
	ctx context.Context,
	environment string,
) (s3InstanceTarget, bool, error) {
	runtime, err := l.runtimes.Runtime(ctx, environment)
	if err != nil {
		return s3InstanceTarget{}, false, rerrors.Wrap(err, "error resolving container runtime")
	}

	listReq := &velez_api.ListSmerds_Request{Environment: environment}

	list, err := runtime.ListContainers(ctx, listReq)
	if err != nil {
		return s3InstanceTarget{}, false, rerrors.Wrap(err, "error listing containers")
	}

	target := s3InstanceTarget{environment: environment, runtime: runtime}

	for i := range list {
		if list[i].Labels[labels.S3InstanceLabel] == l.name {
			target.hasGarage = true
		}

		if list[i].Labels[labels.S3WebUiLabel] == l.name {
			target.webUi = &list[i]
		}
	}

	return target, target.hasGarage || target.webUi != nil, nil
}

type unbindS3BucketOwnersJob struct {
	locator     *s3InstanceLocator
	secrets     secrets.Store
	dataStorage storage.Storage
}

// Best-effort: a stopped or already removed Garage cannot name its owners,
// and that must not block dropping the instance.
func (j *unbindS3BucketOwnersJob) Do(ctx context.Context) error {
	name := j.locator.name

	target, err := j.locator.locate(ctx)
	if err != nil {
		return rerrors.Wrap(err)
	}

	if !target.hasGarage {
		return nil
	}

	adminToken, err := j.secrets.Get(ctx, domain.S3AdminTokenSecretRef(name))
	if err != nil {
		logUnbindSkipped(ctx, name, err)

		return nil
	}

	adminUrl, err := garage.AdminUrl(ctx, target.runtime, domain.S3ServiceName(name))
	if err != nil {
		logUnbindSkipped(ctx, name, err)

		return nil
	}

	keys, err := garage.New(adminUrl, adminToken).ListKeys(ctx)
	if err != nil {
		logUnbindSkipped(ctx, name, err)

		return nil
	}

	for _, key := range keys {
		owner, bucket, isOwnerKey := domain.ParseS3OwnerKeyName(key.Name)
		if !isOwnerKey {
			continue
		}

		err = j.dataStorage.ServiceResources().DeleteResource(ctx, owner, name+"/"+bucket)
		if err != nil {
			log.Ctx(ctx).Warn().
				Str("instance", name).
				Str("bucket", bucket).
				Err(err).
				Msg("error deleting bucket owner binding")
		}
	}

	return nil
}

func logUnbindSkipped(ctx context.Context, name string, err error) {
	log.Ctx(ctx).Warn().
		Str("instance", name).
		Err(err).
		Msg("error reading garage owner keys, owner bindings are left in place")
}

type removeS3WebUiJob struct {
	locator      *s3InstanceLocator
	vervServices service.VervServicesService
}

// The sidecar shares garage's network namespace, so it has to go before garage
// does. A legacy web ui is a separate service.
func (j *removeS3WebUiJob) Do(ctx context.Context) error {
	name := j.locator.name
	webUiName := domain.S3WebUiServiceName(name)

	target, err := j.locator.locate(ctx)
	if err != nil {
		return rerrors.Wrap(err)
	}

	if target.isWebUiSidecar() {
		err = target.runtime.Remove(ctx, webUiName)
		if err != nil {
			return rerrors.Wrap(err, "error removing web ui sidecar "+webUiName)
		}

		return nil
	}

	removeReq := domain.RemoveServiceReq{
		Name:                 webUiName,
		DropRunningInstances: true,
		Environment:          target.environment,
	}

	err = j.vervServices.Remove(ctx, removeReq)
	if err != nil {
		return rerrors.Wrap(err, "error removing web ui service "+webUiName)
	}

	return nil
}

type removeS3ServiceJob struct {
	locator      *s3InstanceLocator
	vervServices service.VervServicesService
}

func (j *removeS3ServiceJob) Do(ctx context.Context) error {
	serviceName := domain.S3ServiceName(j.locator.name)

	target, err := j.locator.locate(ctx)
	if err != nil {
		return rerrors.Wrap(err)
	}

	removeReq := domain.RemoveServiceReq{
		Name:                 serviceName,
		DropRunningInstances: true,
		Environment:          target.environment,
	}

	err = j.vervServices.Remove(ctx, removeReq)
	if err != nil {
		return rerrors.Wrap(err, "error removing service "+serviceName)
	}

	return nil
}

type removeS3DockerResourcesJob struct {
	nodeClients node_clients.NodeClients
	name        string
}

// Mirrors dinds.DropDind: cleanup after the service is gone is best-effort.
func (j *removeS3DockerResourcesJob) Do(ctx context.Context) error {
	cli := j.nodeClients.Docker().Client()

	err := cli.NetworkRemove(ctx, domain.S3NetworkName(j.name))
	if err != nil && !cerrdefs.IsNotFound(err) {
		log.Ctx(ctx).Warn().
			Str("instance", j.name).
			Err(err).
			Msg("error removing s3 network")
	}

	volumes := []string{domain.S3MetaVolumeName(j.name), domain.S3DataVolumeName(j.name)}

	for _, volume := range volumes {
		err = cli.VolumeRemove(ctx, volume, false)
		if err != nil && !cerrdefs.IsNotFound(err) {
			log.Ctx(ctx).Warn().
				Str("instance", j.name).
				Str("volume", volume).
				Err(err).
				Msg("error removing s3 volume")
		}
	}

	return nil
}

type deleteS3SecretsJob struct {
	secrets secrets.Store
	name    string
}

func (j *deleteS3SecretsJob) Do(ctx context.Context) error {
	scope := domain.S3AdminTokenSecretRef(j.name).Scope

	refs, err := j.secrets.ListRefs(ctx, scope, j.name)
	if err != nil {
		return rerrors.Wrap(err, "error listing instance secrets")
	}

	for _, ref := range refs {
		err = j.secrets.Delete(ctx, ref)
		if err != nil && !rerrors.Is(err, user_errors.ErrSecretNotFound) {
			return rerrors.Wrap(err, "error deleting instance secret")
		}
	}

	return nil
}

type deleteS3ConfigEntriesJob struct {
	configResolver service.ServiceConfigResolver
	dataStorage    storage.Storage
	name           string
}

func (j *deleteS3ConfigEntriesJob) Do(ctx context.Context) error {
	dropped := []string{domain.S3ServiceName(j.name), domain.S3WebUiServiceName(j.name)}

	for _, serviceName := range dropped {
		err := j.configResolver.Delete(ctx, serviceName)
		if err != nil {
			return rerrors.Wrap(err, "error deleting service config")
		}

		err = j.dataStorage.ServiceDependencies().DeleteDependenciesOf(ctx, serviceName)
		if err != nil {
			return rerrors.Wrap(err, "error deleting service dependencies")
		}
	}

	return nil
}
