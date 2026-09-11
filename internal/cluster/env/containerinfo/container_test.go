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
