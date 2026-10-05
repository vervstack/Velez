package node_clients

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testTcpHost  = "tcp://192.168.1.44:2375"
	testUnixHost = "unix:///var/run/docker.sock"
)

func Test_DockerPingHint_Cases(t *testing.T) {
	cases := []struct {
		name          string
		goos          string
		daemonHost    string
		isInContainer bool
		want          string
	}{
		{"remote daemon from macos", goosDarwin, testTcpHost, false,
			pingHintRemoteDaemon + "; " + pingHintDarwinLocalNetwork},
		{"remote daemon from windows", goosWindows, testTcpHost, false,
			pingHintRemoteDaemon + "; " + pingHintWindowsFirewall},
		{"remote daemon from linux", goosLinux, testTcpHost, false, pingHintRemoteDaemon},
		{"remote daemon from container", goosLinux, testTcpHost, true, pingHintRemoteDaemon},
		{"local socket in container", goosLinux, testUnixHost, true, pingHintInContainer},
		{"local socket on macos", goosDarwin, testUnixHost, false, pingHintDarwinLocal},
		{"local socket on linux", goosLinux, testUnixHost, false, pingHintLinuxLocal},
		{"local pipe on windows", goosWindows, "npipe:////./pipe/docker_engine", false, pingHintWindowsLocal},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			want := pingHintPrefix + tc.daemonHost + ": " + tc.want
			require.Equal(t, want, dockerPingHint(tc.goos, tc.daemonHost, tc.isInContainer))
		})
	}
}
