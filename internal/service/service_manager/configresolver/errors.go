package configresolver

import (
	"go.redsock.ru/rerrors"
)

var (
	errInvalidEnvKey = rerrors.New("env key cannot be stored in matreshka: " +
		"expected at least 3 symbols of A-Z, 0-9, '_' and '-'")
)
