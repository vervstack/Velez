package domain_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain"
)

const (
	testScope = "pgaas"
	testOwner = "my-pg"
	testKey   = "password"
)

func TestSecretRef_String(t *testing.T) {
	ref := domain.SecretRef{
		Scope: testScope,
		Owner: testOwner,
		Key:   testKey,
	}

	require.Equal(t, "pgaas/my-pg/password", ref.String())
}

func TestParseSecretRef(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		want    domain.SecretRef
		wantErr bool
	}{
		{
			name:  "valid ref",
			input: "pgaas/my-pg/password",
			want: domain.SecretRef{
				Scope: testScope,
				Owner: testOwner,
				Key:   testKey,
			},
		},
		{
			name:  "valid ref with plugin scope",
			input: "plugin/registry/password",
			want: domain.SecretRef{
				Scope: "plugin",
				Owner: "registry",
				Key:   "password",
			},
		},
		{
			name:    "too few segments",
			input:   "pgaas/my-pg",
			wantErr: true,
		},
		{
			name:    "too many segments",
			input:   "pgaas/my-pg/password/extra",
			wantErr: true,
		},
		{
			name:    "empty segment",
			input:   "pgaas//password",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := domain.ParseSecretRef(tc.input)

			if tc.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, domain.ErrInvalidSecretRef)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestParseSecretRef_RoundTripsWithString(t *testing.T) {
	ref := domain.SecretRef{
		Scope: testScope,
		Owner: testOwner,
		Key:   testKey,
	}

	parsed, err := domain.ParseSecretRef(ref.String())
	require.NoError(t, err)
	require.Equal(t, ref, parsed)
}
