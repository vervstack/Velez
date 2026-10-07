//go:build e2e_full

package e2e

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	dockernetwork "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	networkApiIsolationEnv  = "E2ENETAPIOTHER"
	networkApiClusterSuffix = "netapi-cluster"

	networkApiBridgeName = "bridge"
	networkApiExternalIp = "1.1.1.1"

	networkApiExecTimeout = 10 * time.Second
	networkApiExecPoll    = 250 * time.Millisecond
)

func networkApiSuffix(t *testing.T) string {
	t.Helper()

	return dockerSafeToken(t.Name())
}

func newNetworkApiEnvironment(t *testing.T, extraEnvironments ...string) *TestEnvironment {
	t.Helper()

	return Planes[0].NewEnvironment(t,
		WithContainerSuffix(networkApiSuffix(t)),
		WithEnvironments(extraEnvironments))
}

func newCreateNetworkRequest(
	name, environment string, isInternal, isIccEnabled bool,
) *velez_api.CreateNetwork_Request {
	return &velez_api.CreateNetwork_Request{
		Name:         name,
		Environment:  environment,
		IsInternal:   isInternal,
		IsIccEnabled: isIccEnabled,
	}
}

func newListNetworksRequest(environment string, isForeignIncluded bool) *velez_api.ListNetworks_Request {
	return &velez_api.ListNetworks_Request{
		Environment:       environment,
		IsForeignIncluded: isForeignIncluded,
	}
}

func newGetNetworkRequest(id, environment string) *velez_api.GetNetwork_Request {
	return &velez_api.GetNetwork_Request{Id: id, Environment: environment}
}

func newDeleteNetworkRequest(id, environment string) *velez_api.DeleteNetwork_Request {
	return &velez_api.DeleteNetwork_Request{Id: id, Environment: environment}
}

func newConnectContainerRequest(
	networkId, containerName, environment string, aliases ...string,
) *velez_api.ConnectContainer_Request {
	return &velez_api.ConnectContainer_Request{
		NetworkId:     networkId,
		ContainerName: containerName,
		Aliases:       aliases,
		Environment:   environment,
	}
}

func newDisconnectContainerRequest(
	networkId, containerName, environment string,
) *velez_api.DisconnectContainer_Request {
	return &velez_api.DisconnectContainer_Request{
		NetworkId:     networkId,
		ContainerName: containerName,
		Environment:   environment,
	}
}

func newGetNetworkStatusRequest() *velez_api.GetNetworkStatus_Request {
	return &velez_api.GetNetworkStatus_Request{}
}

func newNetworkApiConnection(serviceName, targetNetwork string) *velez_api.Connection {
	return &velez_api.Connection{
		ServiceName:   serviceName,
		TargetNetwork: targetNetwork,
	}
}

func newMakeConnectionsRequest(conn *velez_api.Connection) *velez_api.MakeConnections_Request {
	return &velez_api.MakeConnections_Request{Connections: []*velez_api.Connection{conn}}
}

func newBreakConnectionsRequest(conn *velez_api.Connection) *velez_api.BreakConnections_Request {
	return &velez_api.BreakConnections_Request{Connections: []*velez_api.Connection{conn}}
}

type networkApiContainer struct {
	id         string
	name       string
	dockerName string
}

func startNetworkApiContainer(t *testing.T, env *TestEnvironment, name, suffix string) networkApiContainer {
	t.Helper()

	dockerClient := env.Custom.NodeClients.Docker().Client()

	pullReader, err := dockerClient.ImagePull(t.Context(), NginxAlpineImage, image.PullOptions{})
	require.NoError(t, err)

	var pulled bytes.Buffer

	_, err = pulled.ReadFrom(pullReader)
	require.NoError(t, err)

	err = pullReader.Close()
	require.NoError(t, err)

	dockerName := name
	if suffix != "" {
		dockerName = name + "_" + suffix
	}

	cfg := newNetworkApiContainerConfig(t, suffix)

	t.Cleanup(func() {
		_ = dockerClient.ContainerRemove(context.Background(), dockerName, container.RemoveOptions{Force: true})
	})

	created, err := dockerClient.ContainerCreate(t.Context(), cfg, nil, nil, nil, dockerName)
	require.NoError(t, err)

	err = dockerClient.ContainerStart(t.Context(), created.ID, container.StartOptions{})
	require.NoError(t, err)

	started := networkApiContainer{id: created.ID, name: name, dockerName: dockerName}

	requireWebServerUp(t, env, started)

	return started
}

func newNetworkApiContainerConfig(t *testing.T, suffix string) *container.Config {
	t.Helper()

	return &container.Config{
		Image: NginxAlpineImage,
		Labels: map[string]string{
			testCaseNameLabel:  t.Name(),
			labels.SuffixLabel: suffix,
		},
	}
}

func networkApiExec(t *testing.T, env *TestEnvironment, containerId string, cmd ...string) (int, string) {
	t.Helper()

	dockerClient := env.Custom.NodeClients.Docker().Client()

	execCfg := container.ExecOptions{Cmd: cmd, AttachStdout: true, AttachStderr: true}

	created, err := dockerClient.ContainerExecCreate(t.Context(), containerId, execCfg)
	require.NoError(t, err)

	attached, err := dockerClient.ContainerExecAttach(t.Context(), created.ID, container.ExecStartOptions{})
	require.NoError(t, err)

	defer attached.Close()

	var stdout, stderr bytes.Buffer

	_, err = stdcopy.StdCopy(&stdout, &stderr, attached.Reader)
	require.NoError(t, err)

	inspected, err := dockerClient.ContainerExecInspect(t.Context(), created.ID)
	require.NoError(t, err)

	return inspected.ExitCode, stdout.String() + stderr.String()
}

func requireWebServerUp(t *testing.T, env *TestEnvironment, c networkApiContainer) {
	t.Helper()

	require.Eventually(t, func() bool {
		exitCode, _ := networkApiExec(t, env, c.id, "wget", "-q", "-T", "2", "-O-", "http://127.0.0.1")

		return exitCode == 0
	}, networkApiExecTimeout, networkApiExecPoll, "web server in %q never came up", c.name)
}

func requireReaches(t *testing.T, env *TestEnvironment, from networkApiContainer, target string) {
	t.Helper()

	var lastOutput string

	ok := assertEventually(func() bool {
		exitCode, out := networkApiExec(t, env, from.id, "wget", "-q", "-T", "3", "-O-", "http://"+target)
		lastOutput = out

		return exitCode == 0
	})

	require.True(t, ok, "%q must reach %q, last output: %s", from.name, target, lastOutput)
}

func requireCannotReach(t *testing.T, env *TestEnvironment, from networkApiContainer, target string) {
	t.Helper()

	exitCode, out := networkApiExec(t, env, from.id, "wget", "-q", "-T", "3", "-O-", "http://"+target)
	require.NotZero(t, exitCode, "%q must not reach %q, output: %s", from.name, target, out)
}

func assertEventually(condition func() bool) bool {
	deadline := time.Now().Add(networkApiExecTimeout)

	for time.Now().Before(deadline) {
		if condition() {
			return true
		}

		time.Sleep(networkApiExecPoll)
	}

	return false
}

func createNetwork(t *testing.T, env *TestEnvironment, req *velez_api.CreateNetwork_Request) *velez_api.Network {
	t.Helper()

	resp, err := env.NetworkClient().CreateNetwork(t.Context(), req)
	require.NoError(t, err)

	created := resp.GetNetwork()
	require.NotEmpty(t, created.GetId())

	registerNetworkCleanup(t, env, created.GetId())

	return created
}

func registerNetworkCleanup(t *testing.T, env *TestEnvironment, networkId string) {
	t.Helper()

	dockerClient := env.Custom.NodeClients.Docker().Client()

	t.Cleanup(func() {
		ctx := context.Background()

		inspected, err := dockerClient.NetworkInspect(ctx, networkId, dockernetwork.InspectOptions{})
		if err != nil {
			return
		}

		for memberId := range inspected.Containers {
			_ = dockerClient.NetworkDisconnect(ctx, networkId, memberId, true)
		}

		_ = dockerClient.NetworkRemove(ctx, networkId)
	})
}

func createForeignNetwork(t *testing.T, env *TestEnvironment, name string) string {
	t.Helper()

	dockerClient := env.Custom.NodeClients.Docker().Client()

	createOpts := dockernetwork.CreateOptions{Driver: "bridge"}

	created, err := dockerClient.NetworkCreate(t.Context(), name, createOpts)
	require.NoError(t, err)

	registerNetworkCleanup(t, env, created.ID)

	return created.ID
}

func getNetwork(t *testing.T, env *TestEnvironment, id, environment string) *velez_api.Network {
	t.Helper()

	resp, err := env.NetworkClient().GetNetwork(t.Context(), newGetNetworkRequest(id, environment))
	require.NoError(t, err)

	return resp.GetNetwork()
}

func listNetworks(t *testing.T, env *TestEnvironment, environment string, isForeignIncluded bool) []*velez_api.Network {
	t.Helper()

	req := newListNetworksRequest(environment, isForeignIncluded)

	resp, err := env.NetworkClient().ListNetworks(t.Context(), req)
	require.NoError(t, err)

	return resp.GetNetworks()
}

func findNetworkById(networks []*velez_api.Network, id string) *velez_api.Network {
	for _, n := range networks {
		if n.GetId() == id {
			return n
		}
	}

	return nil
}

func findNetworkByName(networks []*velez_api.Network, name string) *velez_api.Network {
	for _, n := range networks {
		if n.GetName() == name {
			return n
		}
	}

	return nil
}

func connectContainer(t *testing.T, env *TestEnvironment, req *velez_api.ConnectContainer_Request) {
	t.Helper()

	_, err := env.NetworkClient().ConnectContainer(t.Context(), req)
	require.NoError(t, err)
}

func disconnectContainer(t *testing.T, env *TestEnvironment, req *velez_api.DisconnectContainer_Request) {
	t.Helper()

	_, err := env.NetworkClient().DisconnectContainer(t.Context(), req)
	require.NoError(t, err)
}

func requireStatusCode(t *testing.T, err error, want codes.Code) {
	t.Helper()

	require.Error(t, err)
	require.Equal(t, want, status.Code(err), "got: %v", err)
}

func requireMember(t *testing.T, n *velez_api.Network, c networkApiContainer, aliases ...string) {
	t.Helper()

	for _, member := range n.GetMembers() {
		if member.GetContainerId() != c.id {
			continue
		}

		require.Equal(t, c.name, member.GetContainerName())
		require.NotEmpty(t, member.GetIpAddress())

		for _, alias := range aliases {
			require.Contains(t, member.GetAliases(), alias)
		}

		return
	}

	require.Failf(t, "member missing", "container %q is not a member of network %q", c.name, n.GetName())
}

func requireNotMember(t *testing.T, n *velez_api.Network, c networkApiContainer) {
	t.Helper()

	for _, member := range n.GetMembers() {
		require.NotEqual(t, c.id, member.GetContainerId(), "container %q must not be a member", c.name)
	}
}

func managedCapabilities() []velez_api.NetworkCapability {
	return []velez_api.NetworkCapability{
		velez_api.NetworkCapability_NETWORK_CAPABILITY_DELETE,
		velez_api.NetworkCapability_NETWORK_CAPABILITY_ATTACH,
	}
}

func foreignCapabilities() []velez_api.NetworkCapability {
	return []velez_api.NetworkCapability{velez_api.NetworkCapability_NETWORK_CAPABILITY_ATTACH}
}

func dockerProviderCapabilities() []velez_api.NetworkCapability {
	return []velez_api.NetworkCapability{
		velez_api.NetworkCapability_NETWORK_CAPABILITY_CREATE,
		velez_api.NetworkCapability_NETWORK_CAPABILITY_DELETE,
		velez_api.NetworkCapability_NETWORK_CAPABILITY_ATTACH,
		velez_api.NetworkCapability_NETWORK_CAPABILITY_RESTRICT,
	}
}

func inspectDockerNetwork(t *testing.T, env *TestEnvironment, id string) dockernetwork.Inspect {
	t.Helper()

	dockerClient := env.Custom.NodeClients.Docker().Client()

	inspected, err := dockerClient.NetworkInspect(t.Context(), id, dockernetwork.InspectOptions{})
	require.NoError(t, err)

	return inspected
}

func Test_NetworkApi_CreateAndList(t *testing.T) {
	t.Parallel()

	env := newNetworkApiEnvironment(t)

	created := createNetwork(t, env, newCreateNetworkRequest("app_net", "", false, true))

	listed := findNetworkById(listNetworks(t, env, "", false), created.GetId())
	require.NotNil(t, listed, "a created network must be listed")

	require.Equal(t, "app_net", listed.GetName())
	require.True(t, listed.GetIsManaged())
	require.Equal(t, velez_api.NetworkProvider_NETWORK_PROVIDER_DOCKER, listed.GetProvider())
	require.Equal(t, managedCapabilities(), listed.GetCapabilities())

	inspected := inspectDockerNetwork(t, env, created.GetId())
	require.Equal(t, labels.NetworkManagedLabelValue, inspected.Labels[labels.NetworkManagedLabel])
	require.Equal(t, networkApiSuffix(t), inspected.Labels[labels.SuffixLabel])
	require.Equal(t, "app_net_"+networkApiSuffix(t), inspected.Name)

	fetched := getNetwork(t, env, created.GetId(), "")
	require.Equal(t, listed.GetId(), fetched.GetId())
	require.Equal(t, listed.GetName(), fetched.GetName())
	require.Equal(t, listed.GetCapabilities(), fetched.GetCapabilities())
}

func Test_NetworkApi_CreateFlags(t *testing.T) {
	t.Parallel()

	env := newNetworkApiEnvironment(t)

	plain := createNetwork(t, env, newCreateNetworkRequest("plain_net", "", false, true))
	require.True(t, plain.GetIsIccEnabled())
	require.False(t, plain.GetIsInternal())

	plainInspected := inspectDockerNetwork(t, env, plain.GetId())
	require.False(t, plainInspected.Internal)
	require.NotEqual(t, "false", plainInspected.Options["com.docker.network.bridge.enable_icc"])

	restricted := createNetwork(t, env, newCreateNetworkRequest("restricted_net", "", true, false))
	require.True(t, restricted.GetIsInternal())
	require.False(t, restricted.GetIsIccEnabled())

	fetched := getNetwork(t, env, restricted.GetId(), "")
	require.True(t, fetched.GetIsInternal())
	require.False(t, fetched.GetIsIccEnabled())

	restrictedInspected := inspectDockerNetwork(t, env, restricted.GetId())
	require.True(t, restrictedInspected.Internal)
	require.Equal(t, "false", restrictedInspected.Options["com.docker.network.bridge.enable_icc"])
}

func Test_NetworkApi_CreateRejectsDuplicateAndEmptyName(t *testing.T) {
	t.Parallel()

	env := newNetworkApiEnvironment(t)

	createNetwork(t, env, newCreateNetworkRequest("dup_net", "", false, true))

	_, err := env.NetworkClient().CreateNetwork(t.Context(), newCreateNetworkRequest("dup_net", "", false, true))
	requireStatusCode(t, err, codes.AlreadyExists)

	_, err = env.NetworkClient().CreateNetwork(t.Context(), newCreateNetworkRequest("", "", false, true))
	requireStatusCode(t, err, codes.InvalidArgument)

	_, err = env.NetworkClient().CreateNetwork(t.Context(), newCreateNetworkRequest("   ", "", false, true))
	requireStatusCode(t, err, codes.InvalidArgument)
}

func Test_NetworkApi_ConnectAndDns(t *testing.T) {
	t.Parallel()

	suffix := networkApiSuffix(t)
	env := newNetworkApiEnvironment(t)

	appNet := createNetwork(t, env, newCreateNetworkRequest("dns_net", "", false, true))
	server := startNetworkApiContainer(t, env, "srv", suffix)
	client := startNetworkApiContainer(t, env, "cli", suffix)

	connectContainer(t, env, newConnectContainerRequest(appNet.GetId(), server.name, "", "srv-alias"))
	connectContainer(t, env, newConnectContainerRequest(appNet.GetId(), client.name, ""))

	connected := getNetwork(t, env, appNet.GetId(), "")
	require.Len(t, connected.GetMembers(), 2)
	requireMember(t, connected, server, "srv-alias")
	requireMember(t, connected, client)

	requireReaches(t, env, client, "srv-alias")

	disconnectContainer(t, env, newDisconnectContainerRequest(appNet.GetId(), server.name, ""))

	disconnected := getNetwork(t, env, appNet.GetId(), "")
	requireNotMember(t, disconnected, server)
	requireMember(t, disconnected, client)

	requireCannotReach(t, env, client, "srv-alias")
}

func requireIccBehaviour(
	t *testing.T, env *TestEnvironment, suffix, networkName string, isIccEnabled bool,
) {
	t.Helper()

	targetNet := createNetwork(t, env, newCreateNetworkRequest(networkName, "", false, isIccEnabled))
	server := startNetworkApiContainer(t, env, networkName+"_srv", suffix)
	client := startNetworkApiContainer(t, env, networkName+"_cli", suffix)

	alias := networkName + "-srv"

	connectContainer(t, env, newConnectContainerRequest(targetNet.GetId(), server.name, "", alias))
	connectContainer(t, env, newConnectContainerRequest(targetNet.GetId(), client.name, ""))

	if isIccEnabled {
		requireReaches(t, env, client, alias)

		return
	}

	requireCannotReach(t, env, client, alias)
}

func Test_NetworkApi_IccDisabledBlocksTraffic(t *testing.T) {
	t.Parallel()

	suffix := networkApiSuffix(t)
	env := newNetworkApiEnvironment(t)

	requireIccBehaviour(t, env, suffix, "icc_on", true)
	requireIccBehaviour(t, env, suffix, "icc_off", false)
}

func Test_NetworkApi_InternalHasNoEgress(t *testing.T) {
	t.Parallel()

	suffix := networkApiSuffix(t)
	env := newNetworkApiEnvironment(t)

	control := startNetworkApiContainer(t, env, "control", suffix)
	isolated := startNetworkApiContainer(t, env, "isolated", suffix)

	internalNet := createNetwork(t, env, newCreateNetworkRequest("internal_net", "", true, true))
	require.True(t, internalNet.GetIsInternal())

	connectContainer(t, env, newConnectContainerRequest(internalNet.GetId(), isolated.name, ""))

	dockerClient := env.Custom.NodeClients.Docker().Client()

	err := dockerClient.NetworkDisconnect(t.Context(), networkApiBridgeName, isolated.id, true)
	require.NoError(t, err)

	probe := []string{"nc", "-w", "3", "-z", networkApiExternalIp, "80"}

	controlExit, _ := networkApiExec(t, env, control.id, probe...)
	if controlExit != 0 {
		t.Logf("sandbox has no outbound connectivity, asserting the internal flag via docker inspect only")

		inspected := inspectDockerNetwork(t, env, internalNet.GetId())
		require.True(t, inspected.Internal)

		return
	}

	isolatedExit, out := networkApiExec(t, env, isolated.id, probe...)
	require.NotZero(t, isolatedExit, "a container only on an internal network must have no egress, output: %s", out)
}

func Test_NetworkApi_DeleteRules(t *testing.T) {
	t.Parallel()

	suffix := networkApiSuffix(t)
	env := newNetworkApiEnvironment(t)

	appNet := createNetwork(t, env, newCreateNetworkRequest("del_net", "", false, true))
	member := startNetworkApiContainer(t, env, "member", suffix)

	connectContainer(t, env, newConnectContainerRequest(appNet.GetId(), member.name, ""))

	_, err := env.NetworkClient().DeleteNetwork(t.Context(), newDeleteNetworkRequest(appNet.GetId(), ""))
	requireStatusCode(t, err, codes.FailedPrecondition)

	disconnectContainer(t, env, newDisconnectContainerRequest(appNet.GetId(), member.name, ""))

	_, err = env.NetworkClient().DeleteNetwork(t.Context(), newDeleteNetworkRequest(appNet.GetId(), ""))
	require.NoError(t, err)

	_, err = env.NetworkClient().GetNetwork(t.Context(), newGetNetworkRequest(appNet.GetId(), ""))
	requireStatusCode(t, err, codes.NotFound)

	_, err = env.NetworkClient().DeleteNetwork(t.Context(), newDeleteNetworkRequest(appNet.GetId(), ""))
	requireStatusCode(t, err, codes.NotFound)

	_, err = env.NetworkClient().DeleteNetwork(t.Context(), newDeleteNetworkRequest("nonexistent-network-id", ""))
	requireStatusCode(t, err, codes.NotFound)
}

func Test_NetworkApi_ForeignNetworks(t *testing.T) {
	t.Parallel()

	suffix := networkApiSuffix(t)
	env := newNetworkApiEnvironment(t)

	foreignName := "foreign_" + suffix
	foreignId := createForeignNetwork(t, env, foreignName)

	require.Nil(t, findNetworkById(listNetworks(t, env, "", false), foreignId),
		"a foreign network must be hidden by default")

	foreign := findNetworkById(listNetworks(t, env, "", true), foreignId)
	require.NotNil(t, foreign, "a foreign network must be listed when requested")
	require.False(t, foreign.GetIsManaged())
	require.Equal(t, foreignCapabilities(), foreign.GetCapabilities())

	_, err := env.NetworkClient().DeleteNetwork(t.Context(), newDeleteNetworkRequest(foreignId, ""))
	requireStatusCode(t, err, codes.FailedPrecondition)

	member := startNetworkApiContainer(t, env, "member", suffix)

	connectContainer(t, env, newConnectContainerRequest(foreignId, member.name, "", "foreign-alias"))
	requireMember(t, getNetwork(t, env, foreignId, ""), member, "foreign-alias")

	disconnectContainer(t, env, newDisconnectContainerRequest(foreignId, member.name, ""))
	requireNotMember(t, getNetwork(t, env, foreignId, ""), member)

	bridge := findNetworkByName(listNetworks(t, env, "", true), networkApiBridgeName)
	require.NotNil(t, bridge, "the system bridge network must be listed when foreign networks are included")
	require.False(t, bridge.GetIsManaged())
	require.Empty(t, bridge.GetCapabilities())

	_, err = env.NetworkClient().ConnectContainer(t.Context(), newConnectContainerRequest(bridge.GetId(), member.name, ""))
	requireStatusCode(t, err, codes.FailedPrecondition)
}

func Test_NetworkApi_EnvironmentIsolation(t *testing.T) {
	t.Parallel()

	suffix := networkApiSuffix(t)
	otherEnv := networkApiIsolationEnv
	env := newNetworkApiEnvironment(t, otherEnv)

	netA := createNetwork(t, env, newCreateNetworkRequest("iso_a", "", false, true))
	netB := createNetwork(t, env, newCreateNetworkRequest("iso_b", otherEnv, false, true))

	require.Nil(t, findNetworkById(listNetworks(t, env, "", false), netB.GetId()))
	require.Nil(t, findNetworkById(listNetworks(t, env, otherEnv, false), netA.GetId()))
	require.NotNil(t, findNetworkById(listNetworks(t, env, otherEnv, false), netB.GetId()))

	require.Nil(t, findNetworkById(listNetworks(t, env, "", true), netB.GetId()),
		"another environment's managed network must stay hidden even with foreign networks included")

	_, err := env.NetworkClient().GetNetwork(t.Context(), newGetNetworkRequest(netB.GetId(), ""))
	requireStatusCode(t, err, codes.NotFound)

	_, err = env.NetworkClient().GetNetwork(t.Context(), newGetNetworkRequest(netA.GetId(), otherEnv))
	requireStatusCode(t, err, codes.NotFound)

	_, err = env.NetworkClient().DeleteNetwork(t.Context(), newDeleteNetworkRequest(netB.GetId(), ""))
	requireStatusCode(t, err, codes.NotFound)

	_, err = env.NetworkClient().DeleteNetwork(t.Context(), newDeleteNetworkRequest(netA.GetId(), otherEnv))
	requireStatusCode(t, err, codes.NotFound)

	require.NotNil(t, getNetwork(t, env, netA.GetId(), ""), "a refused delete must leave the network in place")
	require.NotNil(t, getNetwork(t, env, netB.GetId(), otherEnv), "a refused delete must leave the network in place")

	containerB := startNetworkApiContainer(t, env, "iso_cont_b", otherEnv)

	_, err = env.NetworkClient().ConnectContainer(t.Context(),
		newConnectContainerRequest(netA.GetId(), containerB.dockerName, ""))
	requireStatusCode(t, err, codes.NotFound)

	connectContainer(t, env, newConnectContainerRequest(netB.GetId(), containerB.name, otherEnv))
	requireMember(t, getNetwork(t, env, netB.GetId(), otherEnv), containerB)

	containerA := startNetworkApiContainer(t, env, "iso_cont_a", suffix)

	_, err = env.NetworkClient().ConnectContainer(t.Context(),
		newConnectContainerRequest(netB.GetId(), containerA.dockerName, otherEnv))
	requireStatusCode(t, err, codes.NotFound)
}

func Test_NetworkApi_LegacyConnectionsStillWork(t *testing.T) {
	t.Parallel()

	suffix := networkApiSuffix(t)
	env := newNetworkApiEnvironment(t)

	appNet := createNetwork(t, env, newCreateNetworkRequest("legacy_net", "", false, true))
	member := startNetworkApiContainer(t, env, "legacy_member", suffix)

	conn := newNetworkApiConnection(member.name, "legacy_net")

	_, err := env.Custom.ApiGrpcImpl.MakeConnections(t.Context(), newMakeConnectionsRequest(conn))
	require.NoError(t, err)

	requireMember(t, getNetwork(t, env, appNet.GetId(), ""), member)

	_, err = env.Custom.ApiGrpcImpl.BreakConnections(t.Context(), newBreakConnectionsRequest(conn))
	require.NoError(t, err)

	requireNotMember(t, getNetwork(t, env, appNet.GetId(), ""), member)
}

func requireDockerOnlyStatus(t *testing.T, resp *velez_api.GetNetworkStatus_Response) {
	t.Helper()

	require.Len(t, resp.GetProviders(), 1)
	require.Equal(t, velez_api.NetworkProvider_NETWORK_PROVIDER_DOCKER, resp.GetProviders()[0].GetProvider())
	require.Equal(t, dockerProviderCapabilities(), resp.GetProviders()[0].GetCapabilities())
}

func getNetworkStatus(t *testing.T, env *TestEnvironment) *velez_api.GetNetworkStatus_Response {
	t.Helper()

	resp, err := env.NetworkClient().GetNetworkStatus(t.Context(), newGetNetworkStatusRequest())
	require.NoError(t, err)

	return resp
}

func Test_NetworkApi_Status(t *testing.T) {
	t.Parallel()

	env := newNetworkApiEnvironment(t)

	resp := getNetworkStatus(t, env)

	require.False(t, resp.GetIsClusterMode())
	require.False(t, resp.GetIsVcnConnected())
	requireDockerOnlyStatus(t, resp)
}

func Test_NetworkApi_StatusWithVcn(t *testing.T) {
	t.Parallel()

	hs := getSharedHeadscale(t)

	env := Planes[0].NewEnvironment(t,
		WithContainerSuffix(networkApiSuffix(t)),
		WithState(t, WithStateVcnEnabled(hs.apiURL, hs.apiKey, hs.loginURL)))

	resp := getNetworkStatus(t, env)

	require.False(t, resp.GetIsClusterMode())
	require.True(t, resp.GetIsVcnConnected())
	require.Len(t, resp.GetProviders(), 2)
	require.Equal(t, velez_api.NetworkProvider_NETWORK_PROVIDER_DOCKER, resp.GetProviders()[0].GetProvider())
	require.Equal(t, dockerProviderCapabilities(), resp.GetProviders()[0].GetCapabilities())
	require.Equal(t, velez_api.NetworkProvider_NETWORK_PROVIDER_VCN, resp.GetProviders()[1].GetProvider())
	require.Empty(t, resp.GetProviders()[1].GetCapabilities())
}

func Test_NetworkApi_StatusClusterMode(t *testing.T) {
	t.Parallel()

	env, _ := enableStatefullPgUnderDind(t, Planes[1], networkApiClusterSuffix)

	resp := getNetworkStatus(t, env)

	require.True(t, resp.GetIsClusterMode())
	requireDockerOnlyStatus(t, resp)
}
