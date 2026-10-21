package local_storage

import (
	"context"

	"github.com/google/uuid"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service/user_errors"
)

// localWhitelist is the single-node/dev storage.WhitelistStorage: Velez is stateless there, so nothing is
// persisted - the single-node allowance is the container's elevated-access label.
type localWhitelist struct{}

func newWhitelistStorage() *localWhitelist {
	return &localWhitelist{}
}

func (l *localWhitelist) List(_ context.Context) ([]domain.WhitelistedImage, error) {
	return nil, nil
}

func (l *localWhitelist) Add(_ context.Context, _ string) (uuid.UUID, bool, error) {
	return uuid.Nil, false, nil
}

func (l *localWhitelist) Remove(_ context.Context, _ uuid.UUID) error {
	return rerrors.Wrap(user_errors.ErrWhitelistNotSupported)
}

func (l *localWhitelist) RemoveByRepository(_ context.Context, _ string) error {
	return rerrors.Wrap(user_errors.ErrWhitelistNotSupported)
}
