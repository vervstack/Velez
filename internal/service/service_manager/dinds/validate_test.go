package dinds

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/user_errors"
)

func Test_ValidateDindName_Scenarios(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantErr error
	}{
		{"simple", "dind", nil},
		{"with digits and dash", "dind-1", nil},
		{"empty", "", user_errors.ErrDindNameInvalid},
		{"uppercase", "Dind", user_errors.ErrDindNameInvalid},
		{"leading dash", "-dind", user_errors.ErrDindNameInvalid},
		{"trailing dash", "dind-", user_errors.ErrDindNameInvalid},
		{"underscore", "dind_1", user_errors.ErrDindNameInvalid},
		{"too long", strings.Repeat("a", maxDindNameLength+1), user_errors.ErrDindNameInvalid},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateDindName(tc.input)

			if tc.wantErr == nil {
				require.NoError(t, err)

				return
			}

			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}
