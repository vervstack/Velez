//go:build e2e_full

package e2e

import (
	"context"
	"errors"
	"io"
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

	requireUpgradeStreamClosesWhenIdle(t, env, svcProxyPlainServiceName)

	_, err := env.ServiceApiClient().SetServiceProxy(
		t.Context(), newSetServiceProxyRequest(svcProxyPlainServiceName, serviceProxyUrl, []string{serviceProxyBypassHost}))
	require.NoError(t, err)

	requireUpgradeStreamEndsDone(t, drainServiceUpgrade(t, env, svcProxyPlainServiceName))

	withProxy := awaitRecreatedContainer(t, env, svcProxyPlainServiceName, initial.ID)

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

func newWatchServiceUpgradeRequest(serviceName string) *velez_api.WatchServiceUpgrade_Request {
	return &velez_api.WatchServiceUpgrade_Request{ServiceName: serviceName}
}

func drainServiceUpgrade(t *testing.T, env *TestEnvironment, serviceName string) []*velez_api.TaskStatus {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), runnerRedeployTimeout)
	defer cancel()

	stream, err := env.TasksApiClient().WatchServiceUpgrade(ctx, newWatchServiceUpgradeRequest(serviceName))
	require.NoError(t, err)

	var messages []*velez_api.TaskStatus

	for {
		message, recvErr := stream.Recv()
		if errors.Is(recvErr, io.EOF) {
			return messages
		}

		require.NoError(t, recvErr)

		messages = append(messages, message)
	}
}

func requireUpgradeStreamClosesWhenIdle(t *testing.T, env *TestEnvironment, serviceName string) {
	t.Helper()

	require.Empty(t, drainServiceUpgrade(t, env, serviceName))
}

func requireUpgradeStreamEndsDone(t *testing.T, messages []*velez_api.TaskStatus) {
	t.Helper()

	require.NotEmpty(t, messages)

	last := messages[len(messages)-1]
	require.Equal(t, velez_api.TaskStatus_DONE, last.GetStatus())

	for _, message := range messages[:len(messages)-1] {
		isInFlight := message.GetStatus() == velez_api.TaskStatus_PENDING ||
			message.GetStatus() == velez_api.TaskStatus_RUNNING
		require.True(t, isInFlight, "non-terminal message had status %s", message.GetStatus())
	}

	hasJobs := false

	for _, message := range messages {
		hasJobs = hasJobs || len(message.GetJobs()) > 0
	}

	require.True(t, hasJobs, "no message carried per-job statuses")
}
