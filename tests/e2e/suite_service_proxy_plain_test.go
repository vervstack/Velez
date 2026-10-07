//go:build e2e_full

package e2e

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/storage/environments"
)

// svcProxyPlainServiceName is >= 4 chars and [A-Za-z0-9_] only, and short: container names double as hostnames.
const svcProxyPlainServiceName = "e2e_svc_proxy_plain"

func newPlainServiceCreateRequest(serviceName string) *velez_api.CreateService_Request {
	return &velez_api.CreateService_Request{
		Name:        serviceName,
		Environment: environments.DefaultEnvironmentName,
	}
}

func newPlainServiceDeployRequest(t *testing.T, serviceName string) *velez_api.CreateDeploy_Request {
	t.Helper()

	spec := &velez_api.CreateSmerd_Request{
		Name:         serviceName,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
		Labels:       map[string]string{testCaseNameLabel: t.Name()},
	}

	return &velez_api.CreateDeploy_Request{
		ServiceName:   serviceName,
		Environment:   environments.DefaultEnvironmentName,
		Specification: &velez_api.CreateDeploy_Request_New{New: spec},
	}
}

// Two consecutive upgrades of one service in one Velez process: the second must not be answered by the finished
// upgrade_smerd task of the first.
func Test_SetServiceProxy_PlainService_SetThenClearRecreatesContainerTwice(t *testing.T) {
	t.Parallel()

	runOnEveryPlane(t, runSetServiceProxyOnPlainService)
}

func runSetServiceProxyOnPlainService(t *testing.T, env *TestEnvironment, _ Plane) {
	initial := deployPlainServiceAndAwaitRunning(t, env, svcProxyPlainServiceName)

	withProxy := setServiceProxyAndAwaitNewContainer(
		t, env, svcProxyPlainServiceName, initial.ID, serviceProxyUrl, []string{serviceProxyBypassHost})

	requireContainerProxyEnv(t, withProxy)

	withoutProxy := setServiceProxyAndAwaitNewContainer(t, env, svcProxyPlainServiceName, withProxy.ID, "", nil)

	requireContainerWithoutProxyEnv(t, withoutProxy)
	requireNonProxyEnvKept(t, withProxy, withoutProxy)
	requireServiceReportsProxy(t, env, svcProxyPlainServiceName, "")
}

func deployPlainServiceAndAwaitRunning(t *testing.T, env *TestEnvironment, serviceName string) container.InspectResponse {
	t.Helper()

	_, err := env.Custom.ServiceApiImpl.CreateService(t.Context(), newPlainServiceCreateRequest(serviceName))
	require.NoError(t, err)

	_, err = env.Custom.ServiceApiImpl.CreateDeploy(t.Context(), newPlainServiceDeployRequest(t, serviceName))
	require.NoError(t, err)

	dockerClient := env.Custom.NodeClients.Docker().Client()

	var running container.InspectResponse

	require.Eventually(t, func() bool {
		current, inspectErr := dockerClient.ContainerInspect(t.Context(), serviceName)
		if inspectErr != nil || current.ContainerJSONBase == nil || current.State == nil {
			return false
		}

		running = current

		return current.State.Running
	}, runnerRedeployTimeout, dindReadyTick, "service container never started")

	return running
}
