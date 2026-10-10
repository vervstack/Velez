package jobs

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	webUiTestInstance = "main"
	webUiTestRootId   = "root123"
	webUiTestImage    = "khairul169/garage-webui:1.1.0"
)

func newTestGarageWebUiSpec() garageWebUiSidecarSpec {
	restart := &velez_api.RestartPolicy{Type: velez_api.RestartPolicyType_unless_stopped}

	secretEnv := map[string]string{
		envWebUiApiAdminKey:  "admin-token",
		envWebUiAuthUserPass: "admin:hash",
	}

	return garageWebUiSidecarSpec{
		containerName:   domain.S3WebUiServiceName(webUiTestInstance),
		serviceName:     domain.S3ServiceName(webUiTestInstance),
		instanceName:    webUiTestInstance,
		imageName:       webUiTestImage,
		rootContainerId: webUiTestRootId,
		restart:         restart,
		env:             garageWebUiEnv("garage", secretEnv),
	}
}

func Test_NewGarageWebUiSidecarRequest_JoinsRootNetwork(t *testing.T) {
	t.Parallel()

	spec := newTestGarageWebUiSpec()

	req := newGarageWebUiSidecarRequest(spec)

	require.Equal(t, spec.containerName, req.ContainerName)
	require.Equal(t, container.NetworkMode("container:"+webUiTestRootId), req.HostConfig.NetworkMode)
	require.Equal(t, webUiTestImage, req.Config.Image)
}

func Test_NewGarageWebUiSidecarRequest_PublishesNothing(t *testing.T) {
	t.Parallel()

	req := newGarageWebUiSidecarRequest(newTestGarageWebUiSpec())

	require.Empty(t, req.HostConfig.PortBindings)
	require.False(t, req.HostConfig.PublishAllPorts)
	require.Empty(t, req.Config.ExposedPorts)
	require.Empty(t, req.Config.Hostname)
	require.Nil(t, req.NetworkingConfig)
}

func Test_NewGarageWebUiSidecarRequest_Labels(t *testing.T) {
	t.Parallel()

	spec := newTestGarageWebUiSpec()

	req := newGarageWebUiSidecarRequest(spec)

	want := map[string]string{
		labels.VervServiceLabel:  spec.serviceName,
		labels.Sidecar:           labelTrueValue,
		labels.S3WebUiLabel:      webUiTestInstance,
		labels.WebUiForLabel:     spec.serviceName,
		labels.WebUiPortLabel:    "3909",
		labels.ComposeGroupLabel: spec.serviceName,
	}

	require.Equal(t, want, req.Config.Labels)
	require.True(t, isRegisteredSidecar(req.Config.Labels, spec.serviceName))
}

func Test_NewGarageWebUiSidecarRequest_Env(t *testing.T) {
	t.Parallel()

	req := newGarageWebUiSidecarRequest(newTestGarageWebUiSpec())

	require.ElementsMatch(t, []string{
		"API_BASE_URL=http://127.0.0.1:3903",
		"S3_ENDPOINT_URL=http://127.0.0.1:3900",
		"S3_REGION=garage",
		"API_ADMIN_KEY=admin-token",
		"AUTH_USER_PASS=admin:hash",
	}, req.Config.Env)
}

func Test_NewGarageWebUiSidecarRequest_RestartPolicy(t *testing.T) {
	t.Parallel()

	req := newGarageWebUiSidecarRequest(newTestGarageWebUiSpec())

	require.Equal(t, container.RestartPolicyOnFailure, req.HostConfig.RestartPolicy.Name)
}
