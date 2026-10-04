package dinds

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
)

const (
	listAllServicesLimit = 1_000_000
)

func (s *Service) ListDinds(ctx context.Context) ([]domain.DindView, error) {
	instances, err := s.dataStorage.DindInstances().ListDindInstances(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing dind instances")
	}

	nameById, err := s.serviceNameById(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	views := make([]domain.DindView, 0, len(instances))

	for _, instance := range instances {
		name, isResolved := nameById[instance.ServiceId]
		if !isResolved {
			continue
		}

		view := domain.DindView{
			DindInstance: instance,
			Name:         name,
			Address:      domain.DindAddress(name),
		}

		views = append(views, view)
	}

	return views, nil
}

// There is no id -> name lookup on storage.ServicesStorage, so every service
// is resolved through GetByName, the same way runneraas.ListRunners joins.
func (s *Service) serviceNameById(ctx context.Context) (map[int64]string, error) {
	listReq := domain.ListServicesReq{
		IncludeInternal: true,
		Paging:          domain.Paging{Limit: listAllServicesLimit},
	}

	list, err := s.vervServices.List(ctx, listReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing services")
	}

	nameById := make(map[int64]string, len(list.Services))

	for _, base := range list.Services {
		svc, getErr := s.dataStorage.Services().GetByName(ctx, base.Name)
		if getErr != nil {
			continue
		}

		nameById[svc.ID] = base.Name
	}

	return nameById, nil
}
