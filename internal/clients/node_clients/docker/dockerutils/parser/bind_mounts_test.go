package parser

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

const (
	testSource = "/host"
	testTarget = "/data"
)

func Test_FromBindMount_PropagationRoundTrip(t *testing.T) {
	for apiValue, dockerValue := range propagationFromApi {
		settings := &velez_api.Container_Settings{
			BindMounts: []*velez_api.BindMount{
				{Source: testSource, Target: testTarget, Propagation: apiValue},
			},
		}

		mounts := FromBindMount(settings)
		require.Len(t, mounts, 1)
		require.Equal(t, mount.TypeBind, mounts[0].Type)
		require.Equal(t, dockerValue, mounts[0].BindOptions.Propagation)

		points := []container.MountPoint{{
			Type:        mount.TypeBind,
			Source:      mounts[0].Source,
			Destination: mounts[0].Target,
			RW:          true,
			Propagation: mounts[0].BindOptions.Propagation,
		}}

		back := ToBindMount(points)
		require.Len(t, back, 1)
		require.Equal(t, apiValue, back[0].GetPropagation())
	}
}

func Test_FromBindMount_UnspecifiedPropagationHasNoBindOptions(t *testing.T) {
	settings := &velez_api.Container_Settings{
		BindMounts: []*velez_api.BindMount{{Source: testSource, Target: testTarget}},
	}

	mounts := FromBindMount(settings)
	require.Len(t, mounts, 1)
	require.Nil(t, mounts[0].BindOptions)
}

func Test_BindMount_ReadOnlyFlag(t *testing.T) {
	settings := &velez_api.Container_Settings{
		BindMounts: []*velez_api.BindMount{{Source: testSource, Target: testTarget, IsReadOnly: true}},
	}

	mounts := FromBindMount(settings)
	require.True(t, mounts[0].ReadOnly)

	points := []container.MountPoint{{Type: mount.TypeBind, Source: testSource, Destination: testTarget, RW: false}}
	back := ToBindMount(points)
	require.True(t, back[0].GetIsReadOnly())
}

func Test_BindMount_NilAndEmpty(t *testing.T) {
	require.Nil(t, FromBindMount(nil))
	require.Nil(t, FromBindMount(&velez_api.Container_Settings{}))
	require.Nil(t, ToBindMount(nil))

	volumesOnly := []container.MountPoint{{Type: mount.TypeVolume, Name: "v", Destination: "/v"}}
	require.Nil(t, ToBindMount(volumesOnly))
}

func Test_ToElevatedAccess_NothingElevated(t *testing.T) {
	require.Nil(t, ToElevatedAccess(nil))
	require.Nil(t, ToElevatedAccess(&container.HostConfig{}))
}

func Test_ElevatedAccess_RoundTrip(t *testing.T) {
	access := &velez_api.ElevatedAccess{
		IsPrivileged: true,
		CapAdd:       []string{"NET_ADMIN", "SYS_PTRACE"},
		Devices: []*velez_api.DeviceMapping{
			{HostPath: "/dev/fuse", ContainerPath: "/dev/fuse", Permissions: "rwm"},
		},
		IsHostPid:    true,
		IsHostIpc:    true,
		IsHostUserns: true,
	}

	hostConfig := &container.HostConfig{}
	ApplyElevatedAccess(hostConfig, access)

	require.True(t, hostConfig.Privileged)
	require.True(t, hostConfig.PidMode.IsHost())
	require.True(t, hostConfig.IpcMode.IsHost())
	require.True(t, hostConfig.UsernsMode.IsHost())

	back := ToElevatedAccess(hostConfig)
	require.Equal(t, access.GetIsPrivileged(), back.GetIsPrivileged())
	require.Equal(t, access.GetCapAdd(), back.GetCapAdd())
	require.Len(t, back.GetDevices(), 1)
	require.Equal(t, "/dev/fuse", back.GetDevices()[0].GetHostPath())
	require.Equal(t, "rwm", back.GetDevices()[0].GetPermissions())
	require.True(t, back.GetIsHostPid())
	require.True(t, back.GetIsHostIpc())
	require.True(t, back.GetIsHostUserns())
}

func Test_ApplyElevatedAccess_NeverClearsPrivileged(t *testing.T) {
	hostConfig := &container.HostConfig{Privileged: true}

	ApplyElevatedAccess(hostConfig, &velez_api.ElevatedAccess{})
	require.True(t, hostConfig.Privileged)

	ApplyElevatedAccess(hostConfig, nil)
	require.True(t, hostConfig.Privileged)

	ApplyElevatedAccess(nil, &velez_api.ElevatedAccess{IsPrivileged: true})
}
