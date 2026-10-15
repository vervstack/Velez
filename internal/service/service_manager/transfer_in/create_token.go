package transfer_in

import (
	"context"
	"time"

	"go.redsock.ru/rerrors"
)

func (s *Service) CreateToken(_ context.Context) (string, time.Time, error) {
	return "", time.Time{}, rerrors.Wrap(errNotImplemented)
}
