package transfer_in

import (
	"encoding/json"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func newInspectJson(t *testing.T, mounts ...container.MountPoint) []byte {
	t.Helper()

	inspect := container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{ID: "abc", HostConfig: &container.HostConfig{}},
		Config:            &container.Config{Image: "nginx"},
		Mounts:            mounts,
	}

	raw, err := json.Marshal(inspect)
	require.NoError(t, err)

	return raw
}

func newContainerManifest(name, owner string, inspectJson []byte) *velez_api.Transfer_ContainerManifest {
	return &velez_api.Transfer_ContainerManifest{Name: name, NetworkOwner: owner, InspectJson: inspectJson}
}

func newItemManifest(name string, kind velez_api.TransferItemKind) *velez_api.Transfer_ItemManifest {
	return &velez_api.Transfer_ItemManifest{Name: name, Kind: kind, TotalBytes: 10}
}

func Test_ParseManifest_AcceptsRootWithSidecarAndItems(t *testing.T) {
	inspectJson := newInspectJson(t)

	begin := &velez_api.Transfer_Begin{
		Containers: []*velez_api.Transfer_ContainerManifest{
			newContainerManifest("web", "", inspectJson),
			newContainerManifest("web-sidecar", "web", inspectJson),
		},
		Items: []*velez_api.Transfer_ItemManifest{
			newItemManifest("nginx", velez_api.TransferItemKind_TRANSFER_ITEM_KIND_IMAGE),
			newItemManifest("data", velez_api.TransferItemKind_TRANSFER_ITEM_KIND_VOLUME),
		},
	}

	manifest, err := parseManifest(begin)
	require.NoError(t, err)

	require.Len(t, manifest.roots(), 1)
	require.Len(t, manifest.sidecars(), 1)
	require.Equal(t, int64(20), manifest.totalBytes)
}

func Test_ParseManifest_RejectsInvalidInput(t *testing.T) {
	inspectJson := newInspectJson(t)
	bindMount := newInspectJson(t, container.MountPoint{Type: mount.TypeBind, Source: "/host", Destination: "/data"})
	imageKind := velez_api.TransferItemKind_TRANSFER_ITEM_KIND_IMAGE

	cases := []struct {
		name  string
		begin *velez_api.Transfer_Begin
		want  error
	}{
		{
			name:  "no containers",
			begin: &velez_api.Transfer_Begin{},
			want:  errManifestEmpty,
		},
		{
			name: "bind mount",
			begin: &velez_api.Transfer_Begin{Containers: []*velez_api.Transfer_ContainerManifest{
				newContainerManifest("web", "", bindMount),
			}},
			want: errBindMountUnsupported,
		},
		{
			name: "broken inspect",
			begin: &velez_api.Transfer_Begin{Containers: []*velez_api.Transfer_ContainerManifest{
				newContainerManifest("web", "", []byte("{")),
			}},
			want: errInvalidInspect,
		},
		{
			name: "duplicate container",
			begin: &velez_api.Transfer_Begin{Containers: []*velez_api.Transfer_ContainerManifest{
				newContainerManifest("web", "", inspectJson),
				newContainerManifest("web", "", inspectJson),
			}},
			want: errDuplicateContainer,
		},
		{
			name: "sidecar owner is not a root",
			begin: &velez_api.Transfer_Begin{Containers: []*velez_api.Transfer_ContainerManifest{
				newContainerManifest("web", "", inspectJson),
				newContainerManifest("side", "ghost", inspectJson),
			}},
			want: errUnknownNetworkOwner,
		},
		{
			name: "only sidecars",
			begin: &velez_api.Transfer_Begin{Containers: []*velez_api.Transfer_ContainerManifest{
				newContainerManifest("side", "web", inspectJson),
			}},
			want: errNoRootContainer,
		},
		{
			name: "duplicate item",
			begin: &velez_api.Transfer_Begin{
				Containers: []*velez_api.Transfer_ContainerManifest{newContainerManifest("web", "", inspectJson)},
				Items: []*velez_api.Transfer_ItemManifest{
					newItemManifest("nginx", imageKind),
					newItemManifest("nginx", imageKind),
				},
			},
			want: errDuplicateItem,
		},
		{
			name: "unspecified item kind",
			begin: &velez_api.Transfer_Begin{
				Containers: []*velez_api.Transfer_ContainerManifest{newContainerManifest("web", "", inspectJson)},
				Items: []*velez_api.Transfer_ItemManifest{
					newItemManifest("nginx", velez_api.TransferItemKind_TRANSFER_ITEM_KIND_UNSPECIFIED),
				},
			},
			want: errInvalidItem,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseManifest(tc.begin)
			require.ErrorIs(t, err, tc.want)
		})
	}
}
