package verv_services

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/services_queries"
)

type VervService struct {
	servicesStorage storage.ServicesStorage
}

func New(dataStorage storage.Storage) *VervService {
	return &VervService{
		servicesStorage: dataStorage.Services(),
	}
}

func (v VervService) Get(ctx context.Context, r domain.GetServiceReq) (domain.Service, error) {
	if r.Id != nil {
		return v.getById(ctx, *r.Id)
	}

	if r.Name != nil {
		return v.getByName(ctx, *r.Name)
	}

	return domain.Service{}, rerrors.New("id or name is required to find service")
}

func (v VervService) getById(ctx context.Context, id uint64) (domain.Service, error) {
	vService, err := v.servicesStorage.GetById(ctx, int64(id))
	if err != nil {
		return domain.Service{}, rerrors.Wrap(err, "error getting service by id from storage")
	}

	return fromStorageToDomainServiceInfo(vService), nil
}

func (v VervService) getByName(ctx context.Context, name string) (domain.Service, error) {
	vService, err := v.servicesStorage.GetByName(ctx, name)
	if err != nil {
		return domain.Service{}, rerrors.Wrap(err, "error getting service by id from storage")
	}

	return fromStorageToDomainServiceInfo(vService), nil
}

func fromStorageToDomainServiceInfo(vServ services_queries.VelezService) domain.Service {
	return domain.Service{
		Id: uint64(vServ.ID),
		ServiceBasicInfo: domain.ServiceBasicInfo{
			Name: vServ.Name,
		},
	}
}
