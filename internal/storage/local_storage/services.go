package local_storage

import (
	"context"

	"go.vervstack.ru/Velez/internal/domain"
)

type services struct{}

func newServicesStorage() *services {
	return &services{}
}

func (s *services) GetById(_ context.Context, _ int64) (domain.Service, error) {
	return domain.Service{}, nil
}

func (s *services) GetByName(_ context.Context, _ string) (domain.Service, error) {
	return domain.Service{}, nil
}

func (s *services) UpsertService(_ context.Context, _ string) error {
	return nil
}

func (s *services) List(_ context.Context, _ domain.ListServicesReq) (domain.ServiceList, error) {
	return domain.ServiceList{}, nil
}
