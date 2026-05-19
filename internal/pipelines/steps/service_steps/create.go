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

	serviceInfoPtr *domain.ServiceBaseInfo
}

func UpsertServiceState(
	dataStorage storage.Storage,
	serviceInfoPtr *domain.ServiceBaseInfo,
) steps.Step {
	return &upsertServiceState{
		dataStorage.Services(),
		serviceInfoPtr,
	}
}

func (u *upsertServiceState) Do(ctx context.Context) error {
	err := u.servicesStorage.UpsertService(ctx, u.serviceInfoPtr.Name)
	if err != nil {
		return rerrors.Wrap(err, "upsert service info")
	}

	return nil
}
