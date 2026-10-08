package s3aas

import (
	"context"
	"sort"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/jobs"
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

	baseByName, err := s.serviceBaseByName(ctx)
	if err != nil {
		return nil, 0, rerrors.Wrap(err)
	}

	instances := make([]domain.S3Instance, 0, len(refs))

	for _, ref := range refs {
		instance := s.describe(ctx, ref)

		base, isKnown := baseByName[domain.S3ServiceName(ref.name)]
		if isKnown {
			instance.Status = base.Status
			instance.DisplayName = base.DisplayName
		}

		instances = append(instances, instance)
	}

	return instances, total, nil
}

func (s *Service) serviceBaseByName(ctx context.Context) (map[string]domain.ServiceBaseInfo, error) {
	listReq := domain.ListServicesReq{
		IncludeInternal: true,
		Paging:          domain.Paging{Limit: listAllServicesLimit},
	}

	list, err := s.vervServices.List(ctx, listReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing services")
	}

	out := make(map[string]domain.ServiceBaseInfo, len(list.Services))
	for _, base := range list.Services {
		out[base.Name] = base
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
	_, err := s.findInstance(ctx, name)
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

	payload := &velez_api.DropS3InstanceTaskPayload{Name: name}

	_, err = s.jobsEngine.EnqueueReplacing(ctx, name, jobs.DropS3InstanceAction, payload)
	if err != nil {
		return rerrors.Wrap(err, "error enqueuing drop s3 instance task")
	}

	return nil
}
