package transfer_out

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
)

func (s *Service) List(_ context.Context) ([]domain.Transfer, error) {
	return nil, rerrors.Wrap(errNotImplemented)
}
