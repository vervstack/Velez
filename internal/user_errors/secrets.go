package user_errors

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

// ErrSecretNotFound is returned by internal/service/secrets.Store.Get and
// .Delete when no value exists for the given domain.SecretRef.
var ErrSecretNotFound = rerrors.NewUserError("secret not found", codes.NotFound)
