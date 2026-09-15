package pgaas

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
)

// pgListAllServicesLimit is a generous ceiling passed as the Paging.Limit of
// an internal VervServicesService.List call used only to join pg_instances
// rows against live service state (see serviceStateByID) - not a page size
// any caller of ListPgInstances ever sees. storage.ServicesStorage.List
// clamps to min(Limit, total), so Limit must exceed any realistic service
// count to mean "all of them".
const (
	pgListAllServicesLimit = 1_000_000
)

func (s *PgaasService) ListPgInstances(
	ctx context.Context, req domain.ListPgInstancesReq,
) (domain.PgInstanceList, error) {
	rows, err := s.dataStorage.PgInstances().ListPgInstances(ctx)
	if err != nil {
		return domain.PgInstanceList{}, rerrors.Wrap(err, "error listing pg instances")
	}

	total := uint64(len(rows))

	rows = paginate(rows, req.Paging)

	stateByID, err := s.serviceStateByID(ctx)
	if err != nil {
		return domain.PgInstanceList{}, err
	}

	instances := make([]domain.PgInstanceView, 0, len(rows))

	for _, row := range rows {
		base, ok := stateByID[row.ServiceId]
		if !ok {
			// A satellite row whose owning service can't be resolved back to
			// live state - deleted out from under it, or (single-node/dev
			// mode) a backend that never persists a real service id -
			// contributes nothing to the list rather than failing it
			// outright. Mirrors verv_services.List's best-effort enrichment.
			continue
		}

		instances = append(instances, domain.PgInstanceView{
			Name:        base.Name,
			DbName:      row.DbName,
			Username:    row.Username,
			Port:        row.Port,
			Environment: base.Env,
			Status:      base.Status,
			ImageName:   base.ImageName,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
		})
	}

	return domain.PgInstanceList{Total: total, Instances: instances}, nil
}

// serviceStateByID resolves every service's id to its live-enriched
// ServiceBaseInfo (name, status, image, environment - via
// VervServicesService.List's own ListSmerds enrichment), so ListPgInstances
// can join velez.pg_instances rows (keyed by service_id) against it without
// duplicating status/environment/image anywhere. There is no direct
// id -> name lookup on storage.ServicesStorage, so this resolves every
// service's id via GetByName - acceptable for a low-cardinality,
// infrequently-called list endpoint.
func (s *PgaasService) serviceStateByID(ctx context.Context) (map[int64]domain.ServiceBaseInfo, error) {
	list, err := s.vervServices.List(ctx, domain.ListServicesReq{
		IncludeInternal: true,
		Paging:          domain.Paging{Limit: pgListAllServicesLimit},
	})
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing services for pg instance join")
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

// paginate applies req over an already-sorted (by service_id,
// PgInstancesStorage.ListPgInstances's own order) slice - the storage layer
// has no paged pg_instances query, so paging happens here, at the service
// layer, over the full result set.
func paginate(rows []domain.PgInstance, paging domain.Paging) []domain.PgInstance {
	if paging.Offset >= uint64(len(rows)) {
		return nil
	}

	rows = rows[paging.Offset:]

	if paging.Limit > 0 && uint64(len(rows)) > paging.Limit {
		rows = rows[:paging.Limit]
	}

	return rows
}
