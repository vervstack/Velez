package transfer_out

import (
	"context"

	"github.com/google/uuid"
	"go.redsock.ru/rerrors"
)

func (s *Service) Cancel(_ context.Context, _ uuid.UUID) error {
	return rerrors.Wrap(errNotImplemented)
}
