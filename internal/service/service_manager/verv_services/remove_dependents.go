package verv_services

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service/user_errors"
)

const (
	removeResolveServicesLimit = 1_000_000
)

func (v *VervServicesServiceImpl) ensureNoDependents(ctx context.Context, serviceName string) error {
	names, err := v.listDependentNames(ctx, serviceName)
	if err != nil {
		return rerrors.Wrap(err, "error listing dependents of service")
	}

	if len(names) == 0 {
		return nil
	}

	return rerrors.Wrap(user_errors.ErrServiceHasDependents, "used by: "+strings.Join(names, ", "))
}

func (v *VervServicesServiceImpl) listDependentNames(ctx context.Context, serviceName string) ([]string, error) {
	callers, err := v.dataStorage.ServiceDependencies().GetCallers(ctx, serviceName)
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting service callers")
	}

	names := make([]string, 0, len(callers))

	for _, caller := range callers {
		if caller.SourceService == serviceName {
			continue
		}

		names = append(names, caller.SourceService)
	}

	runnerNames, err := v.listRunnerNamesUsingDind(ctx, serviceName)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	names = append(names, runnerNames...)

	slices.Sort(names)

	return slices.Compact(names), nil
}

func (v *VervServicesServiceImpl) listRunnerNamesUsingDind(ctx context.Context, serviceName string) ([]string, error) {
	svc, err := v.dataStorage.Services().GetByName(ctx, serviceName)
	if err != nil {
		if rerrors.Is(err, user_errors.ErrStorageNotFound) {
			return nil, nil
		}

		return nil, rerrors.Wrap(err, "error getting service by name")
	}

	runners, err := v.dataStorage.Runners().ListRunners(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing runners")
	}

	var runnerIds []int64

	for _, runner := range runners {
		if runner.DindServiceId == svc.ID {
			runnerIds = append(runnerIds, runner.ServiceID)
		}
	}

	if len(runnerIds) == 0 {
		return nil, nil
	}

	nameById, err := v.serviceNamesByIds(ctx, runnerIds)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	names := make([]string, 0, len(runnerIds))

	for _, id := range runnerIds {
		name, isResolved := nameById[id]
		if !isResolved {
			name = fmt.Sprintf("runner #%d", id)
		}

		names = append(names, name)
	}

	return names, nil
}

// There is no id -> name lookup on storage.ServicesStorage, so services are
// resolved through GetByName, the same way runneraas.ListRunners joins.
func (v *VervServicesServiceImpl) serviceNamesByIds(ctx context.Context, ids []int64) (map[int64]string, error) {
	listReq := domain.ListServicesReq{
		IncludeInternal: true,
		Paging:          domain.Paging{Limit: removeResolveServicesLimit},
	}

	list, err := v.dataStorage.Services().List(ctx, listReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing services")
	}

	nameById := make(map[int64]string, len(ids))

	for _, base := range list.Services {
		svc, getErr := v.dataStorage.Services().GetByName(ctx, base.Name)
		if getErr != nil {
			continue
		}

		if slices.Contains(ids, svc.ID) {
			nameById[svc.ID] = base.Name
		}
	}

	return nameById, nil
}
