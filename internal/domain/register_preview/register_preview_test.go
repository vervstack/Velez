package register_preview

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service/user_errors"
)

const (
	previewRootId    = "rootid0123456789"
	previewSidecarId = "sidecarid0123456"
	previewService   = "web"
)

func newPreviewRoot() container.InspectResponse {
	return container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{
			ID:         previewRootId,
			Name:       "/app",
			HostConfig: &container.HostConfig{NetworkMode: "bridge"},
		},
		Config: &container.Config{
			Image:  "nginx:1",
			Env:    []string{"A=1", "B=x=y"},
			Labels: map[string]string{"keep": "me", labels.DisplayNameLabel: "old"},
		},
		Mounts: []container.MountPoint{{Destination: "/data", Source: "/srv/data"}},
	}
}

func newPreviewSidecar(networkMode string) container.InspectResponse {
	return container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{
			ID:         previewSidecarId,
			Name:       "/helper",
			HostConfig: &container.HostConfig{NetworkMode: container.NetworkMode(networkMode)},
		},
		Config: &container.Config{
			Image:  "busybox",
			Labels: map[string]string{labels.DisplayNameLabel: "helper"},
		},
	}
}

func previewValueOf(values []*velez_api.PreviewValue, key string) *velez_api.PreviewValue {
	for _, value := range values {
		if value.Key == key {
			return value
		}
	}

	return nil
}

func Test_BuildRegisterPreview_DefaultRolesMirrorSidecarRecreate(t *testing.T) {
	t.Parallel()

	input := RegisterPreviewInput{
		Members:     []container.InspectResponse{newPreviewRoot(), newPreviewSidecar("container:app")},
		ServiceName: previewService,
		RootId:      previewRootId,
		ServiceId:   previewRootId,
	}

	previews, err := BuildRegisterPreview(input)
	require.NoError(t, err)
	require.Len(t, previews, 2)

	root, sidecar := previews[0], previews[1]

	require.True(t, root.IsRoot)
	require.True(t, root.IsService)
	require.Equal(t, "app", root.Name)
	require.Equal(t, "nginx:1", root.ImageName)

	require.Equal(t, "B", root.Env[1].Key)
	require.Equal(t, "x=y", root.Env[1].Value)
	require.Equal(t, velez_api.ValueOrigin_TRANSFERRED, root.Env[1].Origin)

	require.Equal(t, velez_api.ValueOrigin_TRANSFERRED, previewValueOf(root.Labels, "keep").Origin)
	require.Equal(t, velez_api.ValueOrigin_ADDED, previewValueOf(root.Labels, labels.VervServiceLabel).Origin)
	require.Equal(t, velez_api.ValueOrigin_ADDED, previewValueOf(root.Labels, labels.DisplayNameLabel).Origin)
	require.Equal(t, previewService, previewValueOf(root.Labels, labels.DisplayNameLabel).Value)

	onboardedFrom := previewValueOf(root.Labels, labels.OnboardedFromLabel)
	require.Equal(t, previewRootId, onboardedFrom.Value)
	require.Equal(t, velez_api.ValueOrigin_ADDED, onboardedFrom.Origin)

	require.Equal(t, "bridge", root.NetworkMode.Value)
	require.Equal(t, velez_api.ValueOrigin_TRANSFERRED, root.NetworkMode.Origin)

	require.False(t, sidecar.IsRoot)
	require.False(t, sidecar.IsService)
	require.Nil(t, previewValueOf(sidecar.Labels, labels.DisplayNameLabel))
	require.Equal(t, velez_api.ValueOrigin_ADDED, previewValueOf(sidecar.Labels, labels.Sidecar).Origin)
	require.Equal(t, "container:app", sidecar.NetworkMode.Value)
	require.Equal(t, velez_api.ValueOrigin_TRANSFERRED, sidecar.NetworkMode.Origin)
	require.Empty(t, sidecar.Ports)
}

func Test_BuildRegisterPreview_ServiceOtherThanRootSwapsLabels(t *testing.T) {
	t.Parallel()

	input := RegisterPreviewInput{
		Members:     []container.InspectResponse{newPreviewRoot(), newPreviewSidecar("container:" + previewRootId)},
		ServiceName: previewService,
		RootId:      previewRootId,
		ServiceId:   previewSidecarId,
	}

	previews, err := BuildRegisterPreview(input)
	require.NoError(t, err)

	root, service := previews[0], previews[1]

	require.True(t, root.IsRoot)
	require.False(t, root.IsService)
	require.Nil(t, previewValueOf(root.Labels, labels.DisplayNameLabel))
	require.NotNil(t, previewValueOf(root.Labels, labels.Sidecar))

	require.True(t, service.IsService)
	require.Equal(t, previewService, previewValueOf(service.Labels, labels.DisplayNameLabel).Value)
	require.Equal(t, previewSidecarId, previewValueOf(service.Labels, labels.OnboardedFromLabel).Value)
	require.Nil(t, previewValueOf(service.Labels, labels.Sidecar))
}

func Test_BuildRegisterPreview_MemberNotPointingAtRootGetsAddedNetworkMode(t *testing.T) {
	t.Parallel()

	input := RegisterPreviewInput{
		Members:     []container.InspectResponse{newPreviewRoot(), newPreviewSidecar("container:other")},
		ServiceName: previewService,
		RootId:      previewRootId,
		ServiceId:   previewRootId,
	}

	previews, err := BuildRegisterPreview(input)
	require.NoError(t, err)

	require.Equal(t, velez_api.ValueOrigin_ADDED, previews[1].NetworkMode.Origin)
}

func Test_BuildRegisterPreview_RoleOutsideGroupIsRejected(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		rootId    string
		serviceId string
	}{
		{name: "unknown root", rootId: "missing", serviceId: previewRootId},
		{name: "unknown service", rootId: previewRootId, serviceId: "missing"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input := RegisterPreviewInput{
				Members:     []container.InspectResponse{newPreviewRoot()},
				ServiceName: previewService,
				RootId:      tc.rootId,
				ServiceId:   tc.serviceId,
			}

			_, err := BuildRegisterPreview(input)
			require.ErrorIs(t, err, user_errors.ErrPreviewRoleNotInGroup)
		})
	}
}
