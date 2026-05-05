package nodes_service

import (
	"context"

	errors "go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
)

type Service struct {
	storage storage.Storage
}

func NewService(st storage.Storage) *Service {
	return &Service{
		st,
	}
}

func (s *Service) ListNodes(ctx context.Context, req domain.ListNodesReq) (domain.NodesList, error) {
	list, err := s.storage.Nodes().List(ctx, req)
	if err != nil {
		return domain.NodesList{}, errors.Wrap(err)
	}

	return list, nil
}
