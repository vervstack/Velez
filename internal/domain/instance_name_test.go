package domain_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/user_errors"
)

func Test_ValidateInstanceName_Cases(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		isValid bool
	}{
		{"two letters", "ft", true},
		{"letter and digit", "a1", true},
		{"dash and underscore", "my-db_1", true},
		{"max length", strings.Repeat("a", 32), true},
		{"over max length", strings.Repeat("a", 33), false},
		{"empty", "", false},
		{"single char", "a", false},
		{"uppercase", "Foo", false},
		{"leading dash", "-x", false},
		{"leading underscore", "_x", false},
		{"space", "x y", false},
		{"dot", "x.y", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := domain.ValidateInstanceName(tc.input)

			if tc.isValid {
				require.NoError(t, err)

				return
			}

			require.ErrorIs(t, err, user_errors.ErrInvalidInstanceName)
		})
	}
}
