//go:build e2e_full

package e2e

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	linksProxyNamePrefix = "e2e-proxy-"
	linksPlainNamePrefix = "e2e-plain-"
	linksExternalProxy   = "192.168.1.44"
	linksProxyPort       = ":1080"
	linksSocksScheme     = "socks5://"

	resourceTypeDind  = "dind"
	resourceTypeProxy = "proxy"
)

func newGetServiceEnvironmentsRequest(serviceName string) *velez_api.GetServiceEnvironments_Request {
	return &velez_api.GetServiceEnvironments_Request{ServiceName: serviceName}
}

func newManagedServiceRequest(name string) *velez_api.CreateSmerd_Request {
	return &velez_api.CreateSmerd_Request{
		Name:         name,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
		Labels:       map[string]string{labels.VervServiceLabel: name},
	}
}

func linksProxyName(dindName string) string {
	return strings.Replace(dindName, dindE2eNamePrefix, linksProxyNamePrefix, 1)
}

func linksPlainName(t *testing.T) string {
	t.Helper()

	return strings.Replace(dindE2eName(t), dindE2eNamePrefix, linksPlainNamePrefix, 1)
}

func getServiceGraph(t *testing.T, env *TestEnvironment, serviceName string) *velez_api.GetServiceGraph_Response {
	t.Helper()

	resp, err := env.ServiceApiClient().GetServiceGraph(t.Context(), newGetServiceGraphRequest(serviceName))
	require.NoError(t, err)

	return resp
}

func findGraphNode(nodes []*velez_api.ServiceDependencyInfo, name string) *velez_api.ServiceDependencyInfo {
	for _, node := range nodes {
		if node.GetServiceName() == name {
			return node
		}
	}

	return nil
}

func requireGraphDependency(
	t *testing.T, env *TestEnvironment, runnerName, targetName string, nodeType velez_api.NodeType,
) {
	t.Helper()

	graph := getServiceGraph(t, env, runnerName)

	node := findGraphNode(graph.GetDependencies(), targetName)
	require.NotNil(t, node, "%q is not a dependency of %q: %v", targetName, runnerName, graph.GetDependencies())
	require.Equal(t, nodeType, node.GetNodeType())
}

func requireNoGraphDependency(t *testing.T, env *TestEnvironment, runnerName, targetName string) {
	t.Helper()

	graph := getServiceGraph(t, env, runnerName)

	require.Nil(t, findGraphNode(graph.GetDependencies(), targetName))
}

func requireBoundResource(
	t *testing.T, env *TestEnvironment, runnerName, resourceType, wantName string,
) *velez_api.BoundResource {
	t.Helper()

	resources := getServiceResources(t, env, runnerName)

	resource := findBoundResource(resources, resourceType)
	require.NotNil(t, resource, "no %q resource on %q: %v", resourceType, runnerName, resources)
	require.Equal(t, wantName, resource.GetName())

	return resource
}

func requireOnboardedResource(t *testing.T, resource *velez_api.BoundResource, linkedServiceName string) {
	t.Helper()

	require.True(t, resource.GetIsOnboarded(), "%q is not onboarded", resource.GetName())
	require.Equal(t, linkedServiceName, resource.GetLinkedServiceName())
	require.NotEmpty(t, resource.GetContainerId())
}

func requireNotOnboardedResource(t *testing.T, resource *velez_api.BoundResource) {
	t.Helper()

	require.False(t, resource.GetIsOnboarded(), "%q is onboarded", resource.GetName())
	require.Nil(t, resource.LinkedServiceName)
	require.Nil(t, resource.ContainerId)
}

func requireNoProxyResource(t *testing.T, env *TestEnvironment, runnerName string) {
	t.Helper()

	require.Nil(t, findBoundResource(getServiceResources(t, env, runnerName), resourceTypeProxy))
}

func requireGraphCaller(t *testing.T, env *TestEnvironment, dindName, runnerName string) {
	t.Helper()

	graph := getServiceGraph(t, env, dindName)

	caller := findGraphNode(graph.GetCallers(), runnerName)
	require.NotNil(t, caller, "runner %q is not a caller of %q: %v", runnerName, dindName, graph.GetCallers())
	require.Equal(t, velez_api.NodeType_NODE_TYPE_SERVICE, caller.GetNodeType())
}

func setRunnerProxyAndAwait(
	t *testing.T, env *TestEnvironment, runnerName, previousContainerId, proxyHost string,
) string {
	t.Helper()

	proxyUrl := linksSocksScheme + proxyHost + linksProxyPort
	req := newSetServiceProxyRequest(runnerName, proxyUrl, nil)

	var setErr error

	require.Eventually(t, func() bool {
		_, setErr = env.ServiceApiClient().SetServiceProxy(t.Context(), req)

		return setErr == nil
	}, runnerRedeployTimeout, dindReadyTick, "the runner never accepted the proxy update")

	recreated := awaitRecreatedContainer(t, env, runnerName, previousContainerId)

	return recreated.ID
}

func Test_RunnerLinks_DindAndProxyAppearInGraphAndResources(t *testing.T) {
	t.Parallel()

	runOnEveryPlane(t, runRunnerLinks)
}

func runRunnerLinks(t *testing.T, env *TestEnvironment, _ Plane) {
	fixture := createRunnerOnDind(t, env)
	proxyName := linksProxyName(fixture.dindName)

	requireGraphDependency(t, env, fixture.runnerContainer, fixture.dindName, velez_api.NodeType_NODE_TYPE_SERVICE)
	dind := requireBoundResource(t, env, fixture.runnerContainer, resourceTypeDind, fixture.dindName)
	requireOnboardedResource(t, dind, fixture.dindName)
	requireNoProxyResource(t, env, fixture.runnerContainer)
	requireGraphCaller(t, env, fixture.dindName, fixture.runnerContainer)

	env.CreateSmerd(t, newManagedServiceRequest(proxyName))

	internalContainerId := setRunnerProxyAndAwait(t, env, fixture.runnerContainer, fixture.inspected.ID, proxyName)

	requireGraphDependency(t, env, fixture.runnerContainer, fixture.dindName, velez_api.NodeType_NODE_TYPE_SERVICE)
	requireGraphDependency(t, env, fixture.runnerContainer, proxyName, velez_api.NodeType_NODE_TYPE_SERVICE)
	requireBoundResource(t, env, fixture.runnerContainer, resourceTypeDind, fixture.dindName)

	proxy := requireBoundResource(t, env, fixture.runnerContainer, resourceTypeProxy, proxyName)
	require.Equal(t, "running", proxy.GetStatus())
	requireOnboardedResource(t, proxy, proxyName)

	setRunnerProxyAndAwait(t, env, fixture.runnerContainer, internalContainerId, linksExternalProxy)

	requireGraphDependency(t, env, fixture.runnerContainer, linksExternalProxy, velez_api.NodeType_NODE_TYPE_RESOURCE)
	requireNoGraphDependency(t, env, fixture.runnerContainer, proxyName)

	external := requireBoundResource(t, env, fixture.runnerContainer, resourceTypeProxy, linksExternalProxy)
	require.Empty(t, external.GetStatus())
	requireNotOnboardedResource(t, external)

	requireGraphCaller(t, env, fixture.dindName, fixture.runnerContainer)
}

func Test_PlainService_HasNoRunnerLinksAndReportsContainerName(t *testing.T) {
	t.Parallel()

	runOnEveryPlane(t, runPlainServiceWithoutRunnerLinks)
}

func runPlainServiceWithoutRunnerLinks(t *testing.T, env *TestEnvironment, _ Plane) {
	serviceName := linksPlainName(t)

	env.CreateSmerd(t, newManagedServiceRequest(serviceName))

	graph := getServiceGraph(t, env, serviceName)
	require.Empty(t, graph.GetDependencies())
	require.Empty(t, graph.GetCallers())

	require.Empty(t, getServiceResources(t, env, serviceName))

	requireEnvironmentContainerName(t, env, serviceName)
}

func requireEnvironmentContainerName(t *testing.T, env *TestEnvironment, serviceName string) {
	t.Helper()

	req := newGetServiceEnvironmentsRequest(serviceName)

	resp, err := env.ServiceApiClient().GetServiceEnvironments(t.Context(), req)
	require.NoError(t, err)
	require.NotEmpty(t, resp.GetEnvironments())

	for _, environment := range resp.GetEnvironments() {
		require.NotEmpty(t, environment.GetContainerName())
		require.Contains(t, environment.GetContainerName(), serviceName)
	}
}
