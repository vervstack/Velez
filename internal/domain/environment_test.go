package domain_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain"
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
		{"tcp with port", "tcp://10.0.0.5:2375", "10.0.0.5"},
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
