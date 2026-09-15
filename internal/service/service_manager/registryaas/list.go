package registryaas

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
)

const (
	// registryaasListAllServicesLimit mirrors pgaas's pgListAllServicesLimit
	// - a generous ceiling passed as the Paging.Limit of an internal
	// VervServicesService.List call used only to join registry_instances
	// rows against live service state (see serviceStateByID), not a page
	// size any caller of ListRegistryInstances ever sees.
	registryaasListAllServicesLimit = 1_000_000
)

func (s *RegistryaasService) ListRegistryInstances(
	ctx context.Context, req domain.ListRegistryInstancesReq,
) (domain.RegistryInstanceList, error) {
	rows, err := s.dataStorage.RegistryInstances().ListRegistryInstances(ctx)
	if err != nil {
		return domain.RegistryInstanceList{}, rerrors.Wrap(err, "error listing registry instances")
	}

	total := uint64(len(rows))

	rows = paginateRegistryInstances(rows, req.Paging)

	stateByID, err := s.serviceStateByID(ctx)
	if err != nil {
		return domain.RegistryInstanceList{}, err
	}

	instances := make([]domain.RegistryInstanceView, 0, len(rows))

	for _, row := range rows {
		base, ok := stateByID[row.ServiceId]
		if !ok {
			// A satellite row whose owning service can't be resolved back to
			// live state - deleted out from under it, or (single-node/dev
			// mode) a backend that never persists a real service id -
			// contributes nothing to the list rather than failing it
			// outright. Mirrors pgaas.ListPgInstances's same skip.
			continue
		}

		instances = append(instances, domain.RegistryInstanceView{
			Name:        base.Name,
			Port:        row.Port,
			UiPort:      row.UiPort,
			Username:    row.Username,
			Environment: base.Env,
			Status:      base.Status,
			ImageName:   base.ImageName,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
		})
	}

	return domain.RegistryInstanceList{Total: total, Instances: instances}, nil
}

// serviceStateByID mirrors pgaas.PgaasService.serviceStateByID exactly - see
// that method's doc comment.
func (s *RegistryaasService) serviceStateByID(ctx context.Context) (map[int64]domain.ServiceBaseInfo, error) {
	list, err := s.vervServices.List(ctx, domain.ListServicesReq{
		IncludeInternal: true,
		Paging:          domain.Paging{Limit: registryaasListAllServicesLimit},
	})
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing services for registry instance join")
	}

	out := make(map[int64]domain.ServiceBaseInfo, len(list.Services))

	for _, base := range list.Services {
		svc, getErr := s.dataStorage.Services().GetByName(ctx, base.Name)
		if getErr != nil {
			continue
		}

		out[svc.ID] = base
	}

	return out, nil
}

// paginateRegistryInstances mirrors pgaas.paginate - the storage layer has no
// paged registry_instances query, so paging happens here, at the service
// layer, over the full result set.
func paginateRegistryInstances(rows []domain.RegistryInstance, paging domain.Paging) []domain.RegistryInstance {
	if paging.Offset >= uint64(len(rows)) {
		return nil
	}

	rows = rows[paging.Offset:]

	if paging.Limit > 0 && uint64(len(rows)) > paging.Limit {
		rows = rows[:paging.Limit]
	}

	return rows
}
