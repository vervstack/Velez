//go:build e2e_full

package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/storage/environments"
)

const (
	previewUnknownRoleId = "does-not-exist"
	previewSecretEnvKey  = "TS_AUTHKEY"
	previewSecretEnvData = "supersecret-value"
	previewPlainEnvKey   = "PREVIEW_PLAIN"
	previewPlainEnvData  = "plain-value"
)

func newPreviewRegisterContainerRequest(
	containerId, serviceName, serviceContainerId string,
) *velez_api.PreviewRegisterContainer_Request {
	return &velez_api.PreviewRegisterContainer_Request{
		ContainerId:        containerId,
		Environment:        environments.DefaultEnvironmentName,
		ServiceName:        serviceName,
		ServiceContainerId: serviceContainerId,
		Pattern: &velez_api.PreviewRegisterContainer_Request_Generic{
			Generic: &velez_api.RegisterContainer_Request_GenericPattern{},
		},
	}
}

func previewRegisterContainer(
	t *testing.T, env *TestEnvironment, req *velez_api.PreviewRegisterContainer_Request,
) *velez_api.PreviewRegisterContainer_Response {
	t.Helper()

	resp, err := env.VelezApiClient().PreviewRegisterContainer(t.Context(), req)
	require.NoError(t, err)

	return resp
}

func startPreviewGroup(
	t *testing.T, env *TestEnvironment, prefix string,
) (root, sidecar foreignContainer, serviceName string) {
	t.Helper()

	rootName, sidecarName, serviceName := networkGroupNames(prefix)

	root = startForeignContainer(t, env, rootName, false)
	sidecar = startNetworkSidecar(t, env, sidecarName, root.id)

	return root, sidecar, serviceName
}

func requirePreviewOf(
	t *testing.T, resp *velez_api.PreviewRegisterContainer_Response, containerId string,
) *velez_api.ContainerPreview {
	t.Helper()

	for _, preview := range resp.GetContainers() {
		if preview.GetContainerId() == containerId {
			return preview
		}
	}

	require.Failf(t, "container missing from preview", "container %q is not in %+v", containerId, resp.GetContainers())

	return nil
}

func previewValueOf(values []*velez_api.PreviewValue, key string) *velez_api.PreviewValue {
	for _, value := range values {
		if value.GetKey() == key {
			return value
		}
	}

	return nil
}

func requirePreviewLabel(
	t *testing.T, preview *velez_api.ContainerPreview, key, value string, origin velez_api.ValueOrigin,
) {
	t.Helper()

	got := previewValueOf(preview.GetLabels(), key)
	require.NotNil(t, got, "label %q missing from the preview of %q", key, preview.GetName())
	require.Equal(t, value, got.GetValue(), "label %q of %q", key, preview.GetName())
	require.Equal(t, origin, got.GetOrigin(), "origin of label %q of %q", key, preview.GetName())
}

func requireNoPreviewLabel(t *testing.T, preview *velez_api.ContainerPreview, key string) {
	t.Helper()

	require.Nil(t, previewValueOf(preview.GetLabels(), key), "label %q must not be on %q", key, preview.GetName())
}

func requirePreviewServiceLabels(
	t *testing.T, preview *velez_api.ContainerPreview, serviceName, onboardedFromId string,
) {
	t.Helper()

	requirePreviewLabel(t, preview, labels.VervServiceLabel, serviceName, velez_api.ValueOrigin_ADDED)
	requirePreviewLabel(t, preview, labels.DisplayNameLabel, serviceName, velez_api.ValueOrigin_ADDED)
	requirePreviewOnboardedFrom(t, preview, onboardedFromId)
	requirePreviewLabel(t, preview, testCaseNameLabel, t.Name(), velez_api.ValueOrigin_TRANSFERRED)
	requireNoPreviewLabel(t, preview, labels.Sidecar)
}

// requirePreviewOnboardedFrom expects the label to name the replaced container; only the recreated root
// carries it, so an empty id expects no label.
func requirePreviewOnboardedFrom(t *testing.T, preview *velez_api.ContainerPreview, onboardedFromId string) {
	t.Helper()

	if onboardedFromId == "" {
		requireNoPreviewLabel(t, preview, labels.OnboardedFromLabel)

		return
	}

	requirePreviewLabel(t, preview, labels.OnboardedFromLabel, onboardedFromId, velez_api.ValueOrigin_ADDED)
}

func requirePreviewSidecarLabels(
	t *testing.T, preview *velez_api.ContainerPreview, serviceName, onboardedFromId string,
) {
	t.Helper()

	requirePreviewLabel(t, preview, labels.VervServiceLabel, serviceName, velez_api.ValueOrigin_ADDED)
	requirePreviewLabel(t, preview, labels.Sidecar, labelValueTrue, velez_api.ValueOrigin_ADDED)
	requirePreviewLabel(t, preview, testCaseNameLabel, t.Name(), velez_api.ValueOrigin_TRANSFERRED)
	requireNoPreviewLabel(t, preview, labels.DisplayNameLabel)
	requirePreviewOnboardedFrom(t, preview, onboardedFromId)
}

func requirePreviewJoinsRoot(t *testing.T, preview *velez_api.ContainerPreview, rootName string) {
	t.Helper()

	require.NotNil(t, preview.GetNetworkMode())
	require.Equal(t, networkGroupNetworkPrefix+rootName, preview.GetNetworkMode().GetValue())
}

func requirePreviewMatchesInspected(
	t *testing.T, env *TestEnvironment, preview *velez_api.ContainerPreview,
) {
	t.Helper()

	inspected := inspectContainer(t, env, preview.GetName())

	comparedKeys := []string{
		labels.VervServiceLabel,
		labels.DisplayNameLabel,
		labels.Sidecar,
		labels.OnboardedFromLabel,
		testCaseNameLabel,
	}

	want := make(map[string]string)
	got := make(map[string]string)

	for _, key := range comparedKeys {
		previewed := previewValueOf(preview.GetLabels(), key)
		if previewed != nil {
			want[key] = previewed.GetValue()
		}

		actual, isPresent := inspected.Config.Labels[key]
		if isPresent {
			got[key] = actual
		}
	}

	require.Equal(t, want, got, "preview labels of %q must equal the labels the registration stamps", preview.GetName())
}

func Test_RegisterContainer_Preview_GroupDefaultRoles(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t)

	root, sidecar, serviceName := startPreviewGroup(t, env, "pv_default")

	resp := previewRegisterContainer(t, env, newPreviewRegisterContainerRequest(root.id, serviceName, ""))
	require.Len(t, resp.GetContainers(), 2)

	rootPreview := requirePreviewOf(t, resp, root.id)
	require.True(t, rootPreview.GetIsRoot())
	require.True(t, rootPreview.GetIsService())
	requirePreviewServiceLabels(t, rootPreview, serviceName, root.id)

	sidecarPreview := requirePreviewOf(t, resp, sidecar.id)
	require.False(t, sidecarPreview.GetIsRoot())
	require.False(t, sidecarPreview.GetIsService())
	requirePreviewSidecarLabels(t, sidecarPreview, serviceName, "")
	requirePreviewJoinsRoot(t, sidecarPreview, root.name)
}

func Test_RegisterContainer_Preview_ServiceIsNotRoot(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t)

	root, sidecar, serviceName := startPreviewGroup(t, env, "pv_notroot")

	resp := previewRegisterContainer(t, env, newPreviewRegisterContainerRequest(root.id, serviceName, sidecar.id))
	require.Len(t, resp.GetContainers(), 2)

	sidecarPreview := requirePreviewOf(t, resp, sidecar.id)
	require.False(t, sidecarPreview.GetIsRoot())
	require.True(t, sidecarPreview.GetIsService())
	requirePreviewServiceLabels(t, sidecarPreview, serviceName, "")

	rootPreview := requirePreviewOf(t, resp, root.id)
	require.True(t, rootPreview.GetIsRoot())
	require.False(t, rootPreview.GetIsService())
	requirePreviewSidecarLabels(t, rootPreview, serviceName, root.id)
}

func Test_RegisterContainer_Preview_MatchesRegistration(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t)

	root, sidecar, serviceName := startPreviewGroup(t, env, "pv_match")

	resp := previewRegisterContainer(t, env, newPreviewRegisterContainerRequest(root.id, serviceName, ""))
	require.Len(t, resp.GetContainers(), 2)

	registerNetworkGroupRoot(t, env, root.id, serviceName)

	requirePreviewMatchesInspected(t, env, requirePreviewOf(t, resp, root.id))
	requirePreviewMatchesInspected(t, env, requirePreviewOf(t, resp, sidecar.id))
}

func newPreviewSecretEnv() []string {
	return []string{
		previewSecretEnvKey + "=" + previewSecretEnvData,
		previewPlainEnvKey + "=" + previewPlainEnvData,
	}
}

func requirePreviewEnvMasked(t *testing.T, preview *velez_api.ContainerPreview) {
	t.Helper()

	secret := previewValueOf(preview.GetEnv(), previewSecretEnvKey)
	require.NotNil(t, secret, "env %q missing from the preview of %q", previewSecretEnvKey, preview.GetName())
	require.True(t, secret.GetIsSecret())
	require.Empty(t, secret.GetValue())

	plain := previewValueOf(preview.GetEnv(), previewPlainEnvKey)
	require.NotNil(t, plain, "env %q missing from the preview of %q", previewPlainEnvKey, preview.GetName())
	require.False(t, plain.GetIsSecret())
	require.Equal(t, previewPlainEnvData, plain.GetValue())
}

func Test_RegisterContainer_Preview_MasksSecretEnv(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t)

	rootName, serviceName := registerContainerNames("pv_secret")

	root := startForeignContainerWithEnv(t, env, rootName, false, newPreviewSecretEnv())

	resp := previewRegisterContainer(t, env, newPreviewRegisterContainerRequest(root.id, serviceName, ""))

	requirePreviewEnvMasked(t, requirePreviewOf(t, resp, root.id))
}

func Test_RegisterContainer_Preview_RejectsBadRoles(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t)
	velezClient := env.VelezApiClient()

	root, sidecar, serviceName := startPreviewGroup(t, env, "pv_bad")

	t.Run("unknown service container", func(t *testing.T) {
		req := newPreviewRegisterContainerRequest(root.id, serviceName, previewUnknownRoleId)

		_, err := velezClient.PreviewRegisterContainer(t.Context(), req)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("sidecar as the requested container", func(t *testing.T) {
		req := newPreviewRegisterContainerRequest(sidecar.id, serviceName, "")

		_, err := velezClient.PreviewRegisterContainer(t.Context(), req)
		require.Error(t, err)
		require.NotEqual(t, codes.OK, status.Code(err))
	})
}
