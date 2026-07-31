package verv_services

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

func (v *VervService) List(ctx context.Context, req domain.ListServicesReq) (domain.ServiceList, error) {
	list, err := v.dataStorage.Services().List(ctx, req)
	if err != nil {
		return domain.ServiceList{}, rerrors.Wrap(err, "error listing services")
	}

	// Enrich services with smerd data (best-effort)
	for i := range list.Services {
		err = v.enrichServiceWithSmerdData(ctx, &list.Services[i])
		if err != nil {
			continue
		}
	}

	return list, nil
}

func (v *VervService) enrichServiceWithSmerdData(ctx context.Context, svc *domain.ServiceBaseInfo) error {
	// Query smerds with the service name
	listReq := &velez_api.ListSmerds_Request{
		Name: &svc.Name,
	}

	resp, err := v.containerService.ListSmerds(ctx, listReq)
	if err != nil {
		return rerrors.Wrap(err, "error listing smerds for service")
	}

	if len(resp.GetSmerds()) == 0 {
		// No smerd exists, leave fields empty
		return nil
	}

	// Use the first smerd (latest by creation)
	latestSmerd := resp.GetSmerds()[0]

	svc.ImageName = latestSmerd.GetImageName()

	// Map smerd status to service status (best-effort mapping)
	svc.Status = mapSmerdStatus(latestSmerd.GetStatus())

	// Extract 'env' label if it exists
	if latestSmerd.Labels != nil {
		if envVal, ok := latestSmerd.GetLabels()["env"]; ok {
			svc.Env = envVal
		}

		svc.Repo = latestSmerd.GetLabels()[labels.RepoLabel]
	}

	return nil
}

func mapSmerdStatus(smerdStatus velez_api.Smerd_Status) string {
	switch smerdStatus {
	case velez_api.Smerd_running:
		return StatusRunning
	case velez_api.Smerd_paused:
		return StatusDegraded
	case velez_api.Smerd_exited, velez_api.Smerd_dead:
		return StatusStopped
	case velez_api.Smerd_unknown, velez_api.Smerd_created, velez_api.Smerd_restarting, velez_api.Smerd_removing:
		return ""
	}

	return ""
}
