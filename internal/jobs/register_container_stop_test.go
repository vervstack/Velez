package jobs

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func inspectPublishing(hostPorts ...string) container.InspectResponse {
	bindings := nat.PortMap{}
	for _, p := range hostPorts {
		bindings["80/tcp"] = append(bindings["80/tcp"], nat.PortBinding{HostPort: p})
	}

	return container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{
			HostConfig: &container.HostConfig{PortBindings: bindings},
		},
	}
}

func Test_IsStopOldFirstRequired(t *testing.T) {
	withMounts := inspectPublishing("8080")

	withMounts.Mounts = []container.MountPoint{{Destination: testPathDataDir}}

	cases := []struct {
		name      string
		info      container.InspectResponse
		ports     []*velez_api.Port
		isKeeping bool
		want      bool
	}{
		{name: "new host port and no mounts", info: inspectPublishing("8080"), ports: []*velez_api.Port{hostPort(8081)}},
		{name: "same host port", info: inspectPublishing("8080"), ports: []*velez_api.Port{hostPort(8080)}, want: true},
		{name: "new port with mounts", info: withMounts, ports: []*velez_api.Port{hostPort(8081)}, want: true},
		{name: "keeping ports", info: inspectPublishing("8080"), isKeeping: true, want: true},
		{name: "no ports and no mounts", info: inspectPublishing("8080")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, isStopOldFirstRequired(tc.info, tc.ports, tc.isKeeping))
		})
	}
}
