package transfer_out

import (
	"encoding/json"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/stretchr/testify/require"
)

func newInspect(name, image string, mounts ...container.MountPoint) container.InspectResponse {
	base := &container.ContainerJSONBase{Name: "/" + name}
	cfg := &container.Config{Image: image}

	return container.InspectResponse{ContainerJSONBase: base, Config: cfg, Mounts: mounts}
}

func volumeMount(name string) container.MountPoint {
	return container.MountPoint{Type: mount.TypeVolume, Name: name}
}

func bindMount(source string) container.MountPoint {
	return container.MountPoint{Type: mount.TypeBind, Source: source}
}

func Test_DistinctImages_DedupesInFirstSeenOrder(t *testing.T) {
	containers := []container.InspectResponse{
		newInspect("app", "app:1"),
		newInspect("proxy", "proxy:2"),
		newInspect("app-sidecar", "app:1"),
	}

	require.Equal(t, []string{"app:1", "proxy:2"}, distinctImages(containers))
}

func Test_DistinctVolumes_SkipsBindsAndDedupes(t *testing.T) {
	containers := []container.InspectResponse{
		newInspect("app", "app:1", volumeMount("data"), bindMount("/srv/conf")),
		newInspect("sidecar", "side:1", volumeMount("data"), volumeMount("logs")),
	}

	require.Equal(t, []string{"data", "logs"}, distinctVolumes(containers))
}

func Test_ContainerManifests_RootFirstSidecarsOwnedByRoot(t *testing.T) {
	root := newInspect("app", "app:1")
	sidecars := []container.InspectResponse{newInspect("app-agent", "agent:1")}

	manifests, err := containerManifests(root, sidecars)
	require.NoError(t, err)
	require.Len(t, manifests, 2)

	require.Equal(t, "app", manifests[0].GetName())
	require.Empty(t, manifests[0].GetNetworkOwner())
	require.Equal(t, "app-agent", manifests[1].GetName())
	require.Equal(t, "app", manifests[1].GetNetworkOwner())

	var decoded container.InspectResponse

	err = json.Unmarshal(manifests[1].GetInspectJson(), &decoded)
	require.NoError(t, err)
	require.Equal(t, "agent:1", decoded.Config.Image)
}
