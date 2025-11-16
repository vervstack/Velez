package vpn_manager

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/backservice/headscale"
	"go.vervstack.ru/Velez/internal/clients"
	"go.vervstack.ru/Velez/internal/domain"
)

type Service struct {
	docker clients.Docker
}

func New(docker clients.Docker) *Service {
	return &Service{
		docker: docker,
	}
}

func (s *Service) ListUsers(ctx context.Context) ([]domain.VpnNamespace, error) {
	res, err := s.docker.Exec(ctx, headscale.Name, []string{})
	if err != nil {
		return nil, rerrors.Wrap(err, "")
	}
	_ = res
	return nil, nil
}
