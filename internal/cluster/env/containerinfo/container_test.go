package containerinfo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_GetContainerId(t *testing.T) {
	cases := []struct {
		name            string
		writeDockerEnv  bool
		writeHostname   bool
		hostnameContent string
		wantContainerId string
	}{
		{
			name:            "docker container",
			writeDockerEnv:  true,
			writeHostname:   true,
			hostnameContent: "abc123\n",
			wantContainerId: "abc123",
		},
		{
			name:            "bare linux host has /etc/hostname but no dockerenv marker",
			writeDockerEnv:  false,
			writeHostname:   true,
			hostnameContent: "my-linux-box\n",
			wantContainerId: "",
		},
		{
			name:           "macos dev host has neither file",
			writeDockerEnv: false,
			writeHostname:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()

			dockerEnvPath = filepath.Join(dir, ".dockerenv")
			hostnamePath = filepath.Join(dir, "hostname")
			instanceContainerID = nil

			if tc.writeDockerEnv {
				err := os.WriteFile(dockerEnvPath, nil, 0o644)
				require.NoError(t, err)
			}

			if tc.writeHostname {
				err := os.WriteFile(hostnamePath, []byte(tc.hostnameContent), 0o644)
				require.NoError(t, err)
			}

			id := GetContainerId()

			if tc.wantContainerId == "" {
				require.Nil(t, id)

				return
			}

			require.NotNil(t, id)
			require.Equal(t, tc.wantContainerId, *id)
		})
	}
}

func Test_IsSelf(t *testing.T) {
	dir := t.TempDir()

	dockerEnvPath = filepath.Join(dir, ".dockerenv")
	hostnamePath = filepath.Join(dir, "hostname")
	instanceContainerID = nil

	err := os.WriteFile(dockerEnvPath, nil, 0o644)
	require.NoError(t, err)

	err = os.WriteFile(hostnamePath, []byte("abc123def456\n"), 0o644)
	require.NoError(t, err)

	cases := []struct {
		name        string
		containerId string
		want        bool
	}{
		{"full id extends the short id", "abc123def4560123456789", true},
		{"short id matches itself", "abc123def456", true},
		{"unrelated id", "ffff00001111", false},
		{"empty id", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, IsSelf(tc.containerId))
		})
	}
}

func Test_IsSelf_OutsideContainer(t *testing.T) {
	dir := t.TempDir()

	dockerEnvPath = filepath.Join(dir, ".dockerenv")
	hostnamePath = filepath.Join(dir, "hostname")
	instanceContainerID = nil

	require.False(t, IsSelf("abc123def456"))
}
