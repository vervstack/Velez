package register_preview

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service/user_errors"
)

const (
	previewRootId    = "rootid0123456789"
	previewSidecarId = "sidecarid0123456"
	previewService   = "web"
	previewMissingId = "missing"
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
		if value.GetKey() == key {
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

	require.True(t, root.GetIsRoot())
	require.True(t, root.GetIsService())
	require.Equal(t, "app", root.GetName())
	require.Equal(t, "nginx:1", root.GetImageName())

	require.Equal(t, "B", root.GetEnv()[1].GetKey())
	require.Equal(t, "x=y", root.GetEnv()[1].GetValue())
	require.Equal(t, velez_api.ValueOrigin_TRANSFERRED, root.GetEnv()[1].GetOrigin())

	require.Equal(t, velez_api.ValueOrigin_TRANSFERRED, previewValueOf(root.GetLabels(), "keep").GetOrigin())
	require.Equal(t, velez_api.ValueOrigin_ADDED, previewValueOf(root.GetLabels(), labels.VervServiceLabel).GetOrigin())
	require.Equal(t, velez_api.ValueOrigin_ADDED, previewValueOf(root.GetLabels(), labels.DisplayNameLabel).GetOrigin())
	require.Equal(t, previewService, previewValueOf(root.GetLabels(), labels.DisplayNameLabel).GetValue())

	onboardedFrom := previewValueOf(root.GetLabels(), labels.OnboardedFromLabel)
	require.Equal(t, previewRootId, onboardedFrom.GetValue())
	require.Equal(t, velez_api.ValueOrigin_ADDED, onboardedFrom.GetOrigin())

	require.Equal(t, "bridge", root.GetNetworkMode().GetValue())
	require.Equal(t, velez_api.ValueOrigin_TRANSFERRED, root.GetNetworkMode().GetOrigin())

	require.False(t, sidecar.GetIsRoot())
	require.False(t, sidecar.GetIsService())
	require.Nil(t, previewValueOf(sidecar.GetLabels(), labels.DisplayNameLabel))
	require.Equal(t, velez_api.ValueOrigin_ADDED, previewValueOf(sidecar.GetLabels(), labels.Sidecar).GetOrigin())
	require.Equal(t, "container:app", sidecar.GetNetworkMode().GetValue())
	require.Equal(t, velez_api.ValueOrigin_TRANSFERRED, sidecar.GetNetworkMode().GetOrigin())
	require.Empty(t, sidecar.GetPorts())
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

	require.True(t, root.GetIsRoot())
	require.False(t, root.GetIsService())
	require.Nil(t, previewValueOf(root.GetLabels(), labels.DisplayNameLabel))
	require.NotNil(t, previewValueOf(root.GetLabels(), labels.Sidecar))

	require.True(t, service.GetIsService())
	require.Equal(t, previewService, previewValueOf(service.GetLabels(), labels.DisplayNameLabel).GetValue())
	require.Nil(t, previewValueOf(service.GetLabels(), labels.OnboardedFromLabel))
	require.Nil(t, previewValueOf(service.GetLabels(), labels.Sidecar))

	require.Equal(t, previewRootId, previewValueOf(root.GetLabels(), labels.OnboardedFromLabel).GetValue())
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

	require.Equal(t, velez_api.ValueOrigin_ADDED, previews[1].GetNetworkMode().GetOrigin())
}

func Test_BuildRegisterPreview_RoleOutsideGroupIsRejected(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		rootId    string
		serviceId string
	}{
		{name: "unknown root", rootId: previewMissingId, serviceId: previewRootId},
		{name: "unknown service", rootId: previewRootId, serviceId: previewMissingId},
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

func newSwitchPreviewInput(serviceId string) RegisterPreviewInput {
	oldRoot := newPreviewRoot()

	oldRoot.HostConfig.PortBindings = nat.PortMap{"80/tcp": {{HostIP: "0.0.0.0", HostPort: "8080"}}}

	return RegisterPreviewInput{
		Members:        []container.InspectResponse{oldRoot, newPreviewSidecar("container:app")},
		ServiceName:    previewService,
		RootId:         previewSidecarId,
		ServiceId:      serviceId,
		PreviousRootId: previewRootId,
	}
}

func Test_BuildRegisterPreview_SwitchedRootTakesOverPortsAndNetworkMode(t *testing.T) {
	t.Parallel()

	previews, err := BuildRegisterPreview(newSwitchPreviewInput(previewSidecarId))
	require.NoError(t, err)

	oldRoot, newRoot := previews[0], previews[1]

	require.True(t, newRoot.GetIsRoot())
	require.True(t, newRoot.GetIsService())
	require.Equal(t, "bridge", newRoot.GetNetworkMode().GetValue())
	require.Equal(t, velez_api.ValueOrigin_ADDED, newRoot.GetNetworkMode().GetOrigin())
	require.Len(t, newRoot.GetPorts(), 1)
	require.Equal(t, "80/tcp", newRoot.GetPorts()[0].GetKey())
	require.Equal(t, "0.0.0.0:8080", newRoot.GetPorts()[0].GetValue())
	require.Equal(t, previewRootId, previewValueOf(newRoot.GetLabels(), labels.OnboardedFromLabel).GetValue())
	require.Equal(t, previewService, previewValueOf(newRoot.GetLabels(), labels.DisplayNameLabel).GetValue())

	require.False(t, oldRoot.GetIsRoot())
	require.False(t, oldRoot.GetIsService())
	require.Empty(t, oldRoot.GetPorts())
	require.Equal(t, "container:helper", oldRoot.GetNetworkMode().GetValue())
	require.Equal(t, velez_api.ValueOrigin_ADDED, oldRoot.GetNetworkMode().GetOrigin())
	require.Nil(t, previewValueOf(oldRoot.GetLabels(), labels.OnboardedFromLabel))
	require.Nil(t, previewValueOf(oldRoot.GetLabels(), labels.DisplayNameLabel))
	require.Equal(t, "true", previewValueOf(oldRoot.GetLabels(), labels.Sidecar).GetValue())
}

func Test_BuildRegisterPreview_SwitchedRootAndOldRootAsService(t *testing.T) {
	t.Parallel()

	previews, err := BuildRegisterPreview(newSwitchPreviewInput(previewRootId))
	require.NoError(t, err)

	oldRoot, newRoot := previews[0], previews[1]

	require.True(t, oldRoot.GetIsService())
	require.False(t, newRoot.GetIsService())
	require.Equal(t, previewService, previewValueOf(oldRoot.GetLabels(), labels.DisplayNameLabel).GetValue())
	require.Nil(t, previewValueOf(oldRoot.GetLabels(), labels.Sidecar))
	require.Nil(t, previewValueOf(newRoot.GetLabels(), labels.DisplayNameLabel))
	require.Equal(t, "true", previewValueOf(newRoot.GetLabels(), labels.Sidecar).GetValue())
	require.Equal(t, previewRootId, previewValueOf(newRoot.GetLabels(), labels.OnboardedFromLabel).GetValue())
}

func Test_BuildRegisterPreview_UnknownPreviousRootIsRejected(t *testing.T) {
	t.Parallel()

	input := newSwitchPreviewInput(previewSidecarId)

	input.PreviousRootId = previewMissingId

	_, err := BuildRegisterPreview(input)
	require.ErrorIs(t, err, user_errors.ErrPreviewRoleNotInGroup)
}

func Test_BuildRegisterPreview_MasksSecretEnvValues(t *testing.T) {
	t.Parallel()

	root := newPreviewRoot()

	root.Config.Env = append(root.Config.Env, "TS_AUTHKEY=supersecret-value")

	input := RegisterPreviewInput{
		Members:     []container.InspectResponse{root},
		ServiceName: previewService,
		RootId:      previewRootId,
		ServiceId:   previewRootId,
	}

	previews, err := BuildRegisterPreview(input)
	require.NoError(t, err)
	require.Len(t, previews, 1)

	secret := previewValueOf(previews[0].GetEnv(), "TS_AUTHKEY")
	require.NotNil(t, secret)
	require.Empty(t, secret.GetValue())
	require.True(t, secret.GetIsSecret())
	require.Equal(t, velez_api.ValueOrigin_TRANSFERRED, secret.GetOrigin())

	plain := previewValueOf(previews[0].GetEnv(), "A")
	require.NotNil(t, plain)
	require.Equal(t, "1", plain.GetValue())
	require.False(t, plain.GetIsSecret())
}
