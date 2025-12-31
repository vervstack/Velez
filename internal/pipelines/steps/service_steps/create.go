package service_steps

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/pipelines/steps"
	"go.vervstack.ru/Velez/internal/storage"
)

type upsertServiceState struct {
	servicesStorage storage.ServicesStorage

	serviceInfoPtr *domain.ServiceBasicInfo

	serviceIdRespPtr *uint64
}

func UpsertServiceState(
	dataStorage storage.Storage,
	serviceInfoPtr *domain.ServiceBasicInfo,
	serviceIdRespPtr *uint64,
) steps.Step {
	return &upsertServiceState{
		dataStorage.Services(),
		serviceInfoPtr,
		serviceIdRespPtr,
	}
}

func (u *upsertServiceState) Do(ctx context.Context) error {
	err := u.servicesStorage.UpsertService(ctx, u.serviceInfoPtr.Name)
	if err != nil {
		return rerrors.Wrap(err, "upsert service info")
	}

	service, err := u.servicesStorage.GetByName(ctx, u.serviceInfoPtr.Name)
	if err != nil {
		return rerrors.Wrap(err, "get service info")
	}

	*u.serviceIdRespPtr = uint64(service.ID)
	return nil
}
