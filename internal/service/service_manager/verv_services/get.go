package verv_services

import (
	"context"

	"go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

func (v *VervService) Get(ctx context.Context, r domain.GetServiceReq) (domain.Service, error) {
	if r.Name == "" {
		return domain.Service{}, rerrors.New("name is required to find service")
	}

	service, err := v.dataStorage.Services().GetByName(ctx, r.Name)
	if err != nil {
		return domain.Service{}, rerrors.Wrap(err, "error getting service by name from storage")
	}

	// GetByName backends don't all populate Labels (only the single-node
	// synthetic-velez path does). Derive the core/app classification here so
	// the detail response carries the same labels the service list shows.
	// Resource-type classification is intentionally not done here - it needs
	// the resource-name lookup the list path has, and the detail page only
	// branches on "service-core".
	if len(service.Labels) == 0 {
		service.Labels = domain.ClassifyService(service.Name, "")
	}

	err = v.enrichServiceAbout(ctx, &service)
	if err != nil {
		// best-effort; container may not exist yet
	}

	return service, nil
}

func (v *VervService) enrichServiceAbout(ctx context.Context, svc *domain.Service) error {
	req := &velez_api.ListSmerds_Request{
		Name: &svc.Name,
	}

	resp, err := v.containerService.ListSmerds(ctx, req)
	if err != nil {
		return rerrors.Wrap(err, "error listing smerds for about")
	}

	if len(resp.GetSmerds()) == 0 {
		return nil
	}

	lbl := resp.GetSmerds()[0].GetLabels()
	if lbl == nil {
		return nil
	}

	svc.About = domain.AboutService{
		OriginalName: lbl[labels.VervServiceLabel],
		Description:  lbl[labels.DescriptionLabel],
		Env:          lbl[labels.EnvLabel],
		ServiceType:  lbl[labels.ServiceTypeLabel],
		Team:         lbl[labels.TeamLabel],
		Repo:         lbl[labels.RepoLabel],
		Port:         lbl[labels.PortLabel],
	}

	return nil
}
