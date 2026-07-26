package user_errors

import (
	"go.redsock.ru/rerrors"
)

var (
	ErrNetworkNotFound = rerrors.NewUserError("network not found")
)
