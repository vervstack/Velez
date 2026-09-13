package runneraas

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

// runnerListAllServicesLimit is a generous ceiling passed as the
// Paging.Limit of an internal VervServicesService.List call used only to
// join runners rows against live service state (see serviceStateByID) - not
// a page size any caller of ListRunners ever sees. storage.ServicesStorage
// .List clamps to min(Limit, total), so Limit must exceed any realistic
// service count to mean "all of them".
const (
	runnerListAllServicesLimit = 1_000_000
)

func (s *RunneraasService) ListRunners(ctx context.Context, req domain.ListRunnersReq) (domain.RunnerList, error) {
	rows, err := s.dataStorage.Runners().ListRunners(ctx)
	if err != nil {
		return domain.RunnerList{}, rerrors.Wrap(err, "error listing runners")
	}

	total := uint64(len(rows))

	rows = paginate(rows, req.Paging)

	stateByID, err := s.serviceStateByID(ctx)
	if err != nil {
		return domain.RunnerList{}, err
	}

	runners := make([]domain.RunnerView, 0, len(rows))

	for _, row := range rows {
		base, ok := stateByID[row.ServiceID]
		if !ok {
			// A satellite row whose owning service can't be resolved back to
			// live state - deleted out from under it, or (single-node/dev
			// mode) a backend that never persists a real service id -
			// contributes nothing to the list rather than failing it
			// outright. Mirrors verv_services.List's best-effort enrichment.
			continue
		}

		runners = append(runners, domain.RunnerView{
			Name:        base.Name,
			Provider:    velez_api.RunnerProvider(velez_api.RunnerProvider_value[row.Provider]),
			Scope:       velez_api.RunnerScope(velez_api.RunnerScope_value[row.Scope]),
			Target:      row.Target,
			Labels:      row.Labels,
			Environment: base.Env,
			Status:      base.Status,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
		})
	}

	return domain.RunnerList{Total: total, Runners: runners}, nil
}

// serviceStateByID resolves every service's id to its live-enriched
// ServiceBaseInfo (name, status, environment - via VervServicesService.List's
// own ListSmerds enrichment), so ListRunners can join velez.runners rows
// (keyed by service_id) against it without duplicating status/environment
// anywhere. There is no direct id -> name lookup on storage.ServicesStorage,
// so this resolves every service's id via GetByName - acceptable for a
// low-cardinality, infrequently-called list endpoint.
func (s *RunneraasService) serviceStateByID(ctx context.Context) (map[int64]domain.ServiceBaseInfo, error) {
	list, err := s.vervServices.List(ctx, domain.ListServicesReq{
		IncludeInternal: true,
		Paging:          domain.Paging{Limit: runnerListAllServicesLimit},
	})
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing services for runner join")
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
// RunnersStorage.ListRunners's own order) slice - the storage layer has no
// paged runners query, so paging happens here, at the service layer, over
// the full result set.
func paginate(rows []domain.Runner, paging domain.Paging) []domain.Runner {
	if paging.Offset >= uint64(len(rows)) {
		return nil
	}

	rows = rows[paging.Offset:]

	if paging.Limit > 0 && uint64(len(rows)) > paging.Limit {
		rows = rows[:paging.Limit]
	}

	return rows
}
