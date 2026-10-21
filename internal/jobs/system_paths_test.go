package jobs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_IsSystemPath_Cases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		path string
		want bool
	}{
		{"dev root", "/dev", true},
		{"device node", "/dev/ttyUSB0", true},
		{"proc", "/proc/1/status", true},
		{"sys", "/sys/class/net", true},
		{"run", runRoot, true},
		{"run subdir", "/run/user/1000", true},
		{"var run", "/var/run/docker.sock", true},
		{"docker socket elsewhere", "/srv/docker.sock", true},
		{"unclean path", "/srv/../dev/null", true},
		{"dev prefix is not a segment", "/devices", false},
		{"var runtime is not a segment", "/var/runtime", false},
		{"data directory", "/srv/pgdata", false},
		{"sock inside the name only", "/srv/sockets", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, isSystemPath(tc.path))
		})
	}
}
