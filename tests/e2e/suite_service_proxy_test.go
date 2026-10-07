//go:build e2e_full

package e2e

import (
	"slices"
	"strings"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/proxyenv"
	"go.vervstack.ru/Velez/internal/storage/environments"
)

const (
	serviceProxyUrl        = "socks5://192.168.1.44:1080"
	serviceProxyBypassHost = "gitlab.internal"
)

func newSetServiceProxyRequest(serviceName, proxyUrl string, bypassHosts []string) *velez_api.SetServiceProxy_Request {
	return &velez_api.SetServiceProxy_Request{
		ServiceName:      serviceName,
		Environment:      environments.DefaultEnvironmentName,
		ProxyUrl:         proxyUrl,
		ProxyBypassHosts: bypassHosts,
	}
}

func Test_SetServiceProxy_Runner_RecreatesContainerAndMirrorsJobEnvironment(t *testing.T) {
	t.Parallel()

	runOnEveryPlane(t, runSetServiceProxyOnRunner)
}

func runSetServiceProxyOnRunner(t *testing.T, env *TestEnvironment, _ Plane) {
	fixture := createRunnerOnDind(t, env)

	withProxy := setServiceProxyAndAwaitNewContainer(
		t, env, fixture.runnerContainer, fixture.inspected.ID, serviceProxyUrl, []string{serviceProxyBypassHost})

	requireContainerProxyEnv(t, withProxy)
	requireRunnerJobProxyEnvironment(t, env, fixture.runnerContainer)
	requireServiceReportsProxy(t, env, fixture.runnerContainer, serviceProxyUrl)

	withoutProxy := setServiceProxyAndAwaitNewContainer(t, env, fixture.runnerContainer, withProxy.ID, "", nil)

	requireContainerWithoutProxyEnv(t, withoutProxy)
	requireNonProxyEnvKept(t, withProxy, withoutProxy)
	requireRunnerJobWithoutProxyEnvironment(t, env, fixture.runnerContainer)
	requireServiceReportsProxy(t, env, fixture.runnerContainer, "")
}

func setServiceProxyAndAwaitNewContainer(
	t *testing.T, env *TestEnvironment, serviceName, previousContainerId, proxyUrl string, bypassHosts []string,
) container.InspectResponse {
	t.Helper()

	dockerClient := env.Custom.NodeClients.Docker().Client()

	_, err := env.ServiceApiClient().SetServiceProxy(t.Context(), newSetServiceProxyRequest(serviceName, proxyUrl, bypassHosts))
	require.NoError(t, err)

	var recreated container.InspectResponse

	require.Eventually(t, func() bool {
		current, inspectErr := dockerClient.ContainerInspect(t.Context(), serviceName)
		if inspectErr != nil || current.ContainerJSONBase == nil || current.State == nil {
			return false
		}

		recreated = current

		return current.ID != previousContainerId && current.State.Running
	}, runnerRedeployTimeout, dindReadyTick, "service container was never recreated")

	return recreated
}

func requireContainerProxyEnv(t *testing.T, inspected container.InspectResponse) {
	t.Helper()

	for _, key := range []string{"HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY", "http_proxy", "https_proxy", "all_proxy"} {
		require.Contains(t, inspected.Config.Env, key+"="+serviceProxyUrl)
	}

	noProxy := proxyenv.Env(serviceProxyUrl, []string{serviceProxyBypassHost})["NO_PROXY"]

	require.Contains(t, inspected.Config.Env, "NO_PROXY="+noProxy)
	require.Contains(t, inspected.Config.Env, "no_proxy="+noProxy)
}

func requireContainerWithoutProxyEnv(t *testing.T, inspected container.InspectResponse) {
	t.Helper()

	for _, entry := range inspected.Config.Env {
		for _, key := range proxyenv.Keys() {
			require.False(t, strings.HasPrefix(entry, key+"="), "proxy env %q survived the removal", entry)
		}
	}
}

func requireNonProxyEnvKept(t *testing.T, before, after container.InspectResponse) {
	t.Helper()

	kept := 0

	for _, entry := range before.Config.Env {
		key, _, _ := strings.Cut(entry, "=")
		if slices.Contains(proxyenv.Keys(), key) {
			continue
		}

		require.Contains(t, after.Config.Env, entry, "unrelated env %q was lost by the proxy removal", entry)

		kept++
	}

	require.NotZero(t, kept, "the container carries no unrelated env to prove it survives")
}

// requireRunnerJobProxyEnvironment waits for the config sync, which runs after the container swap
// the caller already awaited.
func requireRunnerJobProxyEnvironment(t *testing.T, env *TestEnvironment, runnerContainer string) {
	t.Helper()

	var config string

	require.Eventually(t, func() bool {
		config = string(readRunnerConfig(t, env, runnerContainer))

		return strings.Contains(config, "environment = [")
	}, runnerRedeployTimeout, dindReadyTick, "config.toml never got a job environment:\n%s", config)

	require.Contains(t, config, `"HTTP_PROXY=`+serviceProxyUrl+`"`)
	require.Contains(t, config, `"http_proxy=`+serviceProxyUrl+`"`)
	require.Contains(t, config, `"https_proxy=`+serviceProxyUrl+`"`)
	require.Contains(t, config, serviceProxyBypassHost+",docker")
	requireRunnerEnvironmentBeforeDockerTable(t, config)
}

func requireRunnerEnvironmentBeforeDockerTable(t *testing.T, config string) {
	t.Helper()

	environmentIndex := strings.Index(config, "environment = [")
	dockerTableIndex := strings.Index(config, "[runners.docker]")

	require.GreaterOrEqual(t, environmentIndex, 0, "config.toml:\n%s", config)
	require.Less(t, environmentIndex, dockerTableIndex, "environment must be a runner-level key:\n%s", config)
}

func requireRunnerJobWithoutProxyEnvironment(t *testing.T, env *TestEnvironment, runnerContainer string) {
	t.Helper()

	var config string

	require.Eventually(t, func() bool {
		config = string(readRunnerConfig(t, env, runnerContainer))

		return !strings.Contains(config, "_PROXY=") && !strings.Contains(config, "_proxy=")
	}, runnerRedeployTimeout, dindReadyTick, "config.toml kept its proxy environment:\n%s", config)

	requireGitlabRunnerDefaults(t, []byte(config), runnerContainer)
}

func requireServiceReportsProxy(t *testing.T, env *TestEnvironment, serviceName, wantProxyUrl string) {
	t.Helper()

	resp, err := env.ServiceApiClient().GetService(t.Context(), newGetServiceRequest(serviceName))
	require.NoError(t, err)

	require.Equal(t, wantProxyUrl, resp.GetProxyUrl())

	if wantProxyUrl == "" {
		require.Empty(t, resp.GetProxyBypassHosts())

		return
	}

	require.Equal(t, []string{serviceProxyBypassHost}, resp.GetProxyBypassHosts())
}
