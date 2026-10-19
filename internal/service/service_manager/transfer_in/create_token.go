package transfer_in

import (
	"context"
	"time"

	"go.redsock.ru/rerrors"
)

func (s *TransferInServiceImpl) CreateToken(ctx context.Context) (string, time.Time, error) {
	token, err := newToken()
	if err != nil {
		return "", time.Time{}, rerrors.Wrap(err, "error generating token")
	}

	expiresAt := tokenExpiry(time.Now())

	err = s.dataStorage.Transfers().CreateToken(ctx, hashToken(token), expiresAt)
	if err != nil {
		return "", time.Time{}, rerrors.Wrap(err, "error storing token")
	}

	return token, expiresAt, nil
}
