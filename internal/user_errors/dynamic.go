package user_errors

import (
	"go.redsock.ru/rerrors"
)

// New builds an error from a message assembled at runtime - a stored task
// or job failure, a formatted diagnostic - rather than a hardcoded string.
// It's the sanctioned path for dynamic error text, mirroring the static
// Err*/err* sentinels elsewhere in this package for hardcoded text. Callers
// must format the message themselves (e.g. fmt.Sprintf) before calling this:
// unlike fmt.Errorf, the args here are rerrors metadata (codes.Code, opts),
// never format verbs.
func New(msg string, args ...any) error {
	return rerrors.New(msg, args...)
}

// NewUserError is New's user-facing counterpart - see New.
func NewUserError(msg string, args ...any) error {
	return rerrors.NewUserError(msg, args...)
}
