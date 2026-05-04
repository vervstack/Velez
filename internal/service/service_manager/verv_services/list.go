package verv_services

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

func (v *VervService) List(ctx context.Context, req domain.ListServicesReq) (domain.ServiceList, error) {
	list, err := v.servicesStorage.List(ctx, req)
	if err != nil {
		return domain.ServiceList{}, rerrors.Wrap(err)
	}

	// Enrich services with smerd data (best-effort)
	for i := range list.Services {
		err = v.enrichServiceWithSmerdData(ctx, &list.Services[i])
		if err != nil {
			// Log the error but continue enriching other services
			// (non-blocking, as smerd may not exist)
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

	if len(resp.Smerds) == 0 {
		// No smerd exists, leave fields empty
		return nil
	}

	// Use the first smerd (latest by creation)
	latestSmerd := resp.Smerds[0]

	svc.ImageName = latestSmerd.ImageName

	// Map smerd status to service status (best-effort mapping)
	svc.Status = mapSmerdStatus(latestSmerd.Status)

	// Extract 'env' label if it exists
	if latestSmerd.Labels != nil {
		if envVal, ok := latestSmerd.Labels["env"]; ok {
			svc.Env = envVal
		}
	}

	return nil
}

func mapSmerdStatus(smerdStatus velez_api.Smerd_Status) string {
	switch smerdStatus {
	case velez_api.Smerd_running:
		return "running"
	case velez_api.Smerd_paused:
		return "degraded"
	case velez_api.Smerd_exited, velez_api.Smerd_dead:
		return "stopped"
	default:
		return ""
	}
}
