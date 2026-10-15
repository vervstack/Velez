package transfer_out

import (
	"context"

	"github.com/google/uuid"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
)

func (s *Service) Start(_ context.Context, _ domain.StartTransferReq) (uuid.UUID, error) {
	return uuid.Nil, rerrors.Wrap(errNotImplemented)
}
