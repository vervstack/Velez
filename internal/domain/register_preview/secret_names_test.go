package register_preview

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_IsSecretName_Cases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		key  string
		want bool
	}{
		{"auth key", "TS_AUTHKEY", true},
		{"password", "POSTGRES_PASSWORD", true},
		{"lowercase token", "api_token", true},
		{"passwd", "DB_PASSWD", true},
		{"secret", "MY_SECRET", true},
		{"hostname", "TS_HOSTNAME", false},
		{"log level", "LOG_LEVEL", false},
		{"port", "PORT", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, isSecretName(tc.key))
		})
	}
}
