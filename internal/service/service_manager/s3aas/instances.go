package s3aas

import (
	"context"
	"sort"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	listAllServicesLimit = 1_000_000
)

func (s *Service) ListInstances(ctx context.Context, paging domain.Paging) ([]domain.S3Instance, uint64, error) {
	refs, err := s.discoverInstances(ctx)
	if err != nil {
		return nil, 0, rerrors.Wrap(err)
	}

	sort.Slice(refs, func(i, j int) bool { return refs[i].name < refs[j].name })

	total := uint64(len(refs))

	refs = paginateRefs(refs, paging)

	statusByName, err := s.statusByName(ctx)
	if err != nil {
		return nil, 0, rerrors.Wrap(err)
	}

	instances := make([]domain.S3Instance, 0, len(refs))

	for _, ref := range refs {
		instance := s.describe(ctx, ref)

		status, isKnown := statusByName[domain.S3ServiceName(ref.name)]
		if isKnown {
			instance.Status = status
		}

		instances = append(instances, instance)
	}

	return instances, total, nil
}

func (s *Service) statusByName(ctx context.Context) (map[string]string, error) {
	listReq := domain.ListServicesReq{
		IncludeInternal: true,
		Paging:          domain.Paging{Limit: listAllServicesLimit},
	}

	list, err := s.vervServices.List(ctx, listReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing services")
	}

	out := make(map[string]string, len(list.Services))
	for _, base := range list.Services {
		out[base.Name] = base.Status
	}

	return out, nil
}

func paginateRefs(refs []instanceRef, paging domain.Paging) []instanceRef {
	if paging.Offset >= uint64(len(refs)) {
		return nil
	}

	refs = refs[paging.Offset:]

	if paging.Limit > 0 && uint64(len(refs)) > paging.Limit {
		refs = refs[:paging.Limit]
	}

	return refs
}

func (s *Service) GetInstanceCredentials(ctx context.Context, name string) (domain.S3InstanceCredentials, error) {
	ref, err := s.findInstance(ctx, name)
	if err != nil {
		return domain.S3InstanceCredentials{}, rerrors.Wrap(err)
	}

	adminToken, err := s.secretsStore.Get(ctx, domain.S3AdminTokenSecretRef(name))
	if err != nil {
		return domain.S3InstanceCredentials{}, rerrors.Wrap(err, "error getting garage admin token")
	}

	instance := s.describe(ctx, ref)

	credentials := domain.S3InstanceCredentials{
		AdminToken:         adminToken,
		S3Endpoint:         s3Endpoint(domain.S3ServiceName(name), instance.RemoteHost, instance.S3Port),
		InternalS3Endpoint: internalEndpoint(domain.S3ServiceName(name), domain.S3ApiContainerPort),
		Region:             instance.Region,
	}

	if ref.webUi == nil {
		return credentials, nil
	}

	password, err := s.secretsStore.Get(ctx, domain.S3WebUiPasswordSecretRef(name))
	if err != nil {
		return domain.S3InstanceCredentials{}, rerrors.Wrap(err, "error getting web ui password")
	}

	credentials.WebUiUrl = webUiEndpoint(
		ref.webUiHostName(), instance.PublishedHost, instance.RemoteHost, instance.WebUiPort,
	)
	credentials.WebUiUsername = domain.S3WebUiUsername
	credentials.WebUiPassword = password

	return credentials, nil
}

func (s *Service) DropInstance(ctx context.Context, name string) error {
	ref, err := s.findInstance(ctx, name)
	if err != nil {
		return rerrors.Wrap(err)
	}

	callers, err := s.dataStorage.ServiceDependencies().GetCallers(ctx, domain.S3ServiceName(name))
	if err != nil {
		return rerrors.Wrap(err, "error getting instance callers")
	}

	if len(callers) > 0 {
		return rerrors.Wrap(user_errors.ErrS3InstanceInUse)
	}

	s.unbindBucketOwners(ctx, ref)

	serviceName := domain.S3ServiceName(name)
	webUiName := domain.S3WebUiServiceName(name)

	if ref.webUi != nil {
		err = s.removeWebUi(ctx, ref, webUiName)
		if err != nil {
			return rerrors.Wrap(err)
		}
	}

	err = s.removeService(ctx, serviceName, ref.environment)
	if err != nil {
		return rerrors.Wrap(err)
	}

	s.removeDockerResources(ctx, name)

	err = s.deleteSecrets(ctx, name)
	if err != nil {
		return rerrors.Wrap(err)
	}

	for _, dropped := range []string{serviceName, webUiName} {
		err = s.configResolver.Delete(ctx, dropped)
		if err != nil {
			return rerrors.Wrap(err, "error deleting service config")
		}

		err = s.dataStorage.ServiceDependencies().DeleteDependenciesOf(ctx, dropped)
		if err != nil {
			return rerrors.Wrap(err, "error deleting service dependencies")
		}
	}

	return nil
}

// removeWebUi drops the web ui sidecar container itself - it is not a
// service, and it shares garage's network namespace, so it has to go before
// garage does. A legacy web ui is a separate service.
func (s *Service) removeWebUi(ctx context.Context, ref instanceRef, webUiName string) error {
	if !ref.isWebUiSidecar() {
		return s.removeService(ctx, webUiName, ref.environment)
	}

	err := ref.runtime.Remove(ctx, webUiName)
	if err != nil {
		return rerrors.Wrap(err, "error removing web ui sidecar "+webUiName)
	}

	return nil
}

func (s *Service) removeService(ctx context.Context, name, environment string) error {
	removeReq := domain.RemoveServiceReq{
		Name:                 name,
		DropRunningInstances: true,
		Environment:          environment,
	}

	err := s.vervServices.Remove(ctx, removeReq)
	if err != nil {
		return rerrors.Wrap(err, "error removing service "+name)
	}

	return nil
}

// Mirrors dinds.DropDind: cleanup after the service is gone is best-effort.
func (s *Service) removeDockerResources(ctx context.Context, name string) {
	err := s.docker.Client().NetworkRemove(ctx, domain.S3NetworkName(name))
	if err != nil && !cerrdefs.IsNotFound(err) {
		log.Ctx(ctx).Warn().
			Str("instance", name).
			Err(err).
			Msg("error removing s3 network")
	}

	volumes := []string{domain.S3MetaVolumeName(name), domain.S3DataVolumeName(name)}

	for _, volume := range volumes {
		err = s.docker.Client().VolumeRemove(ctx, volume, false)
		if err != nil && !cerrdefs.IsNotFound(err) {
			log.Ctx(ctx).Warn().
				Str("instance", name).
				Str("volume", volume).
				Err(err).
				Msg("error removing s3 volume")
		}
	}
}

func (s *Service) deleteSecrets(ctx context.Context, name string) error {
	scope := domain.S3AdminTokenSecretRef(name).Scope

	refs, err := s.secretsStore.ListRefs(ctx, scope, name)
	if err != nil {
		return rerrors.Wrap(err, "error listing instance secrets")
	}

	for _, ref := range refs {
		err = s.secretsStore.Delete(ctx, ref)
		if err != nil && !rerrors.Is(err, user_errors.ErrSecretNotFound) {
			return rerrors.Wrap(err, "error deleting instance secret")
		}
	}

	return nil
}

// Best-effort: a stopped Garage cannot name its owners, and that must not
// block dropping the instance.
func (s *Service) unbindBucketOwners(ctx context.Context, ref instanceRef) {
	_, client, err := s.connect(ctx, ref.name)
	if err != nil {
		log.Ctx(ctx).Warn().
			Str("instance", ref.name).
			Err(err).
			Msg("error connecting to garage, owner bindings are left in place")

		return
	}

	buckets, err := s.listBuckets(ctx, client)
	if err != nil {
		log.Ctx(ctx).Warn().
			Str("instance", ref.name).
			Err(err).
			Msg("error listing buckets, owner bindings are left in place")

		return
	}

	for _, bucket := range buckets {
		if bucket.OwnerService == "" {
			continue
		}

		err = s.unbindOwner(ctx, ref.name, bucket.Name, bucket.OwnerService)
		if err != nil {
			log.Ctx(ctx).Warn().
				Str("instance", ref.name).
				Str("bucket", bucket.Name).
				Err(err).
				Msg("error deleting bucket owner binding")
		}
	}
}
