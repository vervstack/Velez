package version

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Get_Scenarios(t *testing.T) {
	cases := []struct {
		name   string
		baked  string
		wanted string
	}{
		{"baked tag wins", "v1.2.3", "v1.2.3"},
		{"nothing baked is a dev build", "", devVersion},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			version = tc.baked

			t.Cleanup(resetVersion)

			require.Equal(t, tc.wanted, Get())
		})
	}
}

func resetVersion() {
	version = ""
}
