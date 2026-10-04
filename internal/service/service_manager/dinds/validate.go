package dinds

import (
	"regexp"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	maxDindNameLength = 63
)

var dindNamePattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

func validateDindName(name string) error {
	if len(name) > maxDindNameLength || !dindNamePattern.MatchString(name) {
		return rerrors.Wrap(user_errors.ErrDindNameInvalid)
	}

	return nil
}
