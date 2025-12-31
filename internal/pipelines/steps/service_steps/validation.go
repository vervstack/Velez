package service_steps

import (
	"context"

	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"

	"go.vervstack.ru/Velez/internal/pipelines/steps"
)

var allowedNameSymbols = map[rune]struct{}{}

func init() {
	for _, r := range []rune(`ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_abcdefghijklmnopqrstuvwxyz1234567890`) {
		allowedNameSymbols[r] = struct{}{}
	}

}

var (
	ErrInvalidServiceName  = rerrors.New("service name contains invalid characters", codes.InvalidArgument)
	ErrTooShortServiceName = rerrors.New("service name is too short", codes.InvalidArgument)
)

func ValidateServiceName(name string) steps.Step {
	return steps.SingleFunc(func(ctx context.Context) error {

		if len(name) < 4 {
			return rerrors.Wrap(ErrTooShortServiceName)
		}

		invalidCharsMap := map[rune]struct{}{}
		var invalidChars []rune

		for _, r := range []rune(name) {
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
