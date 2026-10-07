package domain_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain"
)

const (
	tcpHostAddress = "10.0.0.5"
	tcpDockerHost  = "tcp://10.0.0.5:2375"
)

func Test_Environment_RemoteHost(t *testing.T) {
	const nodeHost = "unix:///var/run/docker.sock"

	cases := []struct {
		name       string
		dockerHost string
		want       string
	}{
		{"empty is the node itself", "", ""},
		{"node's own daemon", nodeHost, ""},
		{"other local socket", "unix:///run/user/1000/docker.sock", ""},
		{"tcp with port", tcpDockerHost, tcpHostAddress},
		{"tcp hostname", "tcp://build-node:2376", "build-node"},
		{"ssh with user", "ssh://root@node-2", "node-2"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env := domain.Environment{DockerHost: tc.dockerHost}

			require.Equal(t, tc.want, env.RemoteHost(nodeHost))
		})
	}
}

func Test_Environment_PublishedHost(t *testing.T) {
	cases := []struct {
		name       string
		dockerHost string
		nodeHost   string
		want       string
	}{
		{"shared tcp daemon", "", tcpDockerHost, tcpHostAddress},
		{"own daemon equals node tcp daemon", tcpDockerHost, tcpDockerHost, tcpHostAddress},
		{"shared unix socket", "", "unix:///var/run/docker.sock", ""},
		{"other tcp daemon", "tcp://10.0.0.6:2375", tcpDockerHost, "10.0.0.6"},
		{"ssh with user", "ssh://root@node-2", "unix:///var/run/docker.sock", "node-2"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env := domain.Environment{DockerHost: tc.dockerHost}

			require.Equal(t, tc.want, env.PublishedHost(tc.nodeHost))
		})
	}
}
