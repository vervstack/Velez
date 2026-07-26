package service_steps

import (
	"context"

	"go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/pipelines/steps"
	"google.golang.org/grpc/codes"
)

const (
	minServiceNameLen = 4
)

var (
	allowedNameSymbols     = map[rune]struct{}{}
	ErrInvalidServiceName  = rerrors.New("service name contains invalid characters", codes.InvalidArgument)
	ErrTooShortServiceName = rerrors.New("service name is too short", codes.InvalidArgument)
)

func init() {
	const allowedChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_abcdefghijklmnopqrstuvwxyz1234567890"

	for _, r := range allowedChars {
		allowedNameSymbols[r] = struct{}{}
	}
}

func ValidateServiceName(name string) steps.Step {
	return steps.SingleFunc(func(_ context.Context) error {
		if len(name) < minServiceNameLen {
			return rerrors.Wrap(ErrTooShortServiceName)
		}

		invalidCharsMap := map[rune]struct{}{}

		var invalidChars []rune

		for _, r := range name {
			_, ok := allowedNameSymbols[r]
			if !ok {
				_, alreadyHave := invalidCharsMap[r]
				if !alreadyHave {
					invalidCharsMap[r] = struct{}{}
					invalidChars = append(invalidChars, r)
				}
			}
		}

		if len(invalidChars) == 0 {
			return nil
		}

		return rerrors.Wrap(ErrInvalidServiceName, string(invalidChars))
	})
}
