package domain

import (
	"regexp"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/user_errors"
)

var instanceNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{1,31}$`)

func ValidateInstanceName(name string) error {
	if !instanceNamePattern.MatchString(name) {
		return rerrors.Wrap(user_errors.ErrInvalidInstanceName)
	}

	return nil
}
