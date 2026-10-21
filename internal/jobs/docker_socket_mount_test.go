package jobs

import (
	"testing"

	"github.com/docker/docker/api/types/mount"
	"github.com/stretchr/testify/require"
)

func Test_MountsWithDockerSocket_NotGrantedLeavesMountsUntouched(t *testing.T) {
	mounts := []mount.Mount{{Type: mount.TypeBind, Source: "/a", Target: "/b"}}

	got := mountsWithDockerSocket(mounts, false)

	require.Equal(t, mounts, got)
}

func Test_MountsWithDockerSocket_GrantedAppendsSocket(t *testing.T) {
	got := mountsWithDockerSocket(nil, true)

	require.Equal(t, []mount.Mount{dockerSocketMount()}, got)
}

func Test_MountsWithDockerSocket_GrantedSkipsWhenAlreadyMounted(t *testing.T) {
	mounts := []mount.Mount{dockerSocketMount()}

	got := mountsWithDockerSocket(mounts, true)

	require.Len(t, got, 1)
}
