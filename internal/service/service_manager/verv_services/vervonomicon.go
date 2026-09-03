package verv_services

import (
	"context"
	"errors"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	verv "go.vervstack.ru/Velez/internal/domain/vervonomicon"
	"go.vervstack.ru/Velez/internal/service/service_manager/vervonomicon"
)

// GetVervonomicon resolves the .verv/ descriptor a service currently
// deploys from, per docs/features/vervonomicon.md. Absence of a descriptor
// (or of a running instance to read one from) is reported as
// VervonomiconResult.NoDescriptor, never an error - the transport layer
// turns that into a clean empty response.
func (v *VervService) GetVervonomicon(
	ctx context.Context, req domain.GetVervonomiconReq,
) (domain.VervonomiconResult, error) {
	image, err := v.currentServiceImage(ctx, req.ServiceName, req.Environment)
	if err != nil {
		return domain.VervonomiconResult{}, err
	}

	if image == "" {
		return domain.VervonomiconResult{NoDescriptor: true}, nil
	}

	descriptor, files, err := v.readVervonomicon(ctx, req.ServiceName, image, req.Environment)
	if err != nil {
		if errors.Is(err, vervonomicon.ErrNoDescriptor) {
			return domain.VervonomiconResult{NoDescriptor: true}, nil
		}

		return domain.VervonomiconResult{}, err
	}

	resolvedYaml, err := vervonomicon.MarshalResolvedYaml(descriptor)
	if err != nil {
		return domain.VervonomiconResult{}, rerrors.Wrap(err, "error marshalling resolved vervonomicon")
	}

	resourceStatuses, err := v.reconcileResources(ctx, req.ServiceName, descriptor.Resources)
	if err != nil {
		return domain.VervonomiconResult{}, rerrors.Wrap(err, "error reconciling vervonomicon resources")
	}

	result := domain.VervonomiconResult{
		Raw:              files,
		ResolvedYaml:     string(resolvedYaml),
		Source:           descriptor.Source,
		Environment:      req.Environment,
		ResourceStatuses: resourceStatuses,
	}

	return result, nil
}

// currentServiceImage resolves the image a service is currently running in
// environment via ContainerService.ListSmerds - the same live-container-state
// path enrichServiceAbout (get.go) already uses - rather than the stored
// deployment-specification history. ListSmerds is environment-aware and
// uniform across storage backends; by contrast storage.ServicesStorage
// .GetByName never populates ImageName in the postgres backend at all, and
// only does so in local_storage without environment scoping (see
// internal/storage/postgres/services.go vs
// internal/storage/local_storage/services.go). An empty return means the
// service has no running instance in this environment.
func (v *VervService) currentServiceImage(ctx context.Context, serviceName, environment string) (string, error) {
	name := serviceName

	req := &velez_api.ListSmerds_Request{
		Name:        &name,
		Environment: environment,
	}

	resp, err := v.containerService.ListSmerds(ctx, req)
	if err != nil {
		return "", rerrors.Wrap(err, "error listing smerds")
	}

	smerds := resp.GetSmerds()
	if len(smerds) == 0 {
		return "", nil
	}

	return smerds[0].GetImageName(), nil
}

// readVervonomicon reads .verv/ out of image and merges the <environment>/
// overlay. ErrNoDescriptor propagates unwrapped so callers can tell it apart
// from a descriptor that exists but is malformed, via errors.Is.
func (v *VervService) readVervonomicon(
	ctx context.Context, serviceName, image, environment string,
) (verv.Descriptor, map[string][]byte, error) {
	files, err := v.vervSource.Read(ctx, serviceName, image)
	if err != nil {
		return verv.Descriptor{}, nil, rerrors.Wrap(err, "error reading vervonomicon from image")
	}

	descriptor, err := vervonomicon.MergeEnvironment(files, environment)
	if err != nil {
		return verv.Descriptor{}, nil, rerrors.Wrap(err, "error merging vervonomicon environment overlay")
	}

	descriptor.Source = verv.SourceKindImage

	return descriptor, files, nil
}
