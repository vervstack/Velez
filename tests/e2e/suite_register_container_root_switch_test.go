//go:build e2e_full

package e2e

import (
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"go.redsock.ru/toolbox"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service/user_errors"
	"go.vervstack.ru/Velez/internal/storage/environments"
)

const (
	rootSwitchClusterSuffix = registerClusterSuffix + "-rootswitch"
	rootSwitchPort          = nat.Port("80/tcp")
	rootSwitchAlias         = "e2e-root-switch-alias"
	rootSwitchBindSource    = "/tmp/e2e_root_switch"
	rootSwitchRollbackWait  = 30 * time.Second
	rootSwitchRollbackTick  = 500 * time.Millisecond
)

type rootSwitchGroup struct {
	oldRoot     foreignContainer
	newRoot     foreignContainer
	other       foreignContainer
	serviceName string
	networkName string
}

// startRootSwitchGroup runs a network group whose root publishes port 80 and is a whitelisted image, because the
// root rejoins another container's network namespace as a sidecar. With networkName the root sits in that
// user-defined network under rootSwitchAlias.
func startRootSwitchGroup(t *testing.T, env *TestEnvironment, prefix, networkName string) rootSwitchGroup {
	t.Helper()

	rootName, newRootName, serviceName := networkGroupNames(prefix)
	otherName := "e2e_ng_" + prefix + "_other"

	if networkName != "" {
		createRootSwitchNetwork(t, env, networkName)
	}

	oldRoot := startRootSwitchRoot(t, env, rootName, networkName)

	return rootSwitchGroup{
		oldRoot:     oldRoot,
		newRoot:     startNetworkSidecar(t, env, newRootName, oldRoot.id),
		other:       startNetworkSidecar(t, env, otherName, oldRoot.id),
		serviceName: serviceName,
		networkName: networkName,
	}
}

func createRootSwitchNetwork(t *testing.T, env *TestEnvironment, networkName string) {
	t.Helper()

	dockerClient := env.DockerClient()

	_, err := dockerClient.NetworkCreate(t.Context(), networkName, network.CreateOptions{})
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = dockerClient.NetworkRemove(t.Context(), networkName)
	})
}

func startRootSwitchRoot(t *testing.T, env *TestEnvironment, name, networkName string) foreignContainer {
	t.Helper()

	dockerClient := env.DockerClient()

	tagNginxAsSidecarImage(t, env)

	cfg := &container.Config{
		Image:  networkGroupSidecarImage,
		Cmd:    networkGroupSidecarCommand,
		Labels: map[string]string{testCaseNameLabel: t.Name()},
	}

	hostCfg := &container.HostConfig{
		NetworkMode:  container.NetworkMode(networkName),
		PortBindings: nat.PortMap{rootSwitchPort: []nat.PortBinding{{HostIP: "0.0.0.0"}}},
	}

	var networkingCfg *network.NetworkingConfig

	if networkName != "" {
		endpoint := &network.EndpointSettings{Aliases: []string{rootSwitchAlias}}
		networkingCfg = &network.NetworkingConfig{EndpointsConfig: map[string]*network.EndpointSettings{networkName: endpoint}}
	}

	t.Cleanup(func() {
		removeForeignContainer(dockerClient, name, "")
	})

	created, err := dockerClient.ContainerCreate(t.Context(), cfg, hostCfg, networkingCfg, nil, name)
	require.NoError(t, err)

	err = dockerClient.ContainerStart(t.Context(), created.ID, container.StartOptions{})
	require.NoError(t, err)

	return foreignContainer{id: created.ID, name: name}
}

func newRootSwitchRegisterRequest(
	rootId, serviceName, newRootId, serviceContainerId string,
) *velez_api.RegisterContainer_Request {
	req := newRegisterContainerRequest(rootId, serviceName)
	req.RootContainerId = newRootId
	req.ServiceContainerId = serviceContainerId

	return req
}

func registerRootSwitch(t *testing.T, env *TestEnvironment, req *velez_api.RegisterContainer_Request) {
	t.Helper()

	resp := registerContainer(t, env, req)

	task := awaitRegisterTask(t, env, resp)
	require.Equal(t, velez_api.TaskStatus_DONE, task.GetStatus(), "register task error: %s", task.GetError())
}

func requireNoContainerNamed(t *testing.T, env *TestEnvironment, name string) {
	t.Helper()

	nameFilter := filters.NewArgs(filters.Arg("name", "^/"+name+"$"))

	listed, err := env.DockerClient().ContainerList(t.Context(), container.ListOptions{All: true, Filters: nameFilter})
	require.NoError(t, err)
	require.Empty(t, listed, "container %q must be gone", name)
}

func requirePublishedPort(t *testing.T, inspected container.InspectResponse) string {
	t.Helper()

	bindings := inspected.NetworkSettings.Ports[rootSwitchPort]
	require.NotEmpty(t, bindings, "port %s must be published", rootSwitchPort)

	return bindings[0].HostPort
}

func requireNewRootOwnsNetwork(
	t *testing.T, env *TestEnvironment, group rootSwitchGroup, oldRootPort string,
) container.InspectResponse {
	t.Helper()

	newRoot := inspectContainer(t, env, group.newRoot.name)

	require.NotEqual(t, group.newRoot.id, newRoot.ID, "the new root is recreated")
	require.True(t, newRoot.State.Running, "the new root must run")
	require.Equal(t, group.oldRoot.id, newRoot.Config.Labels[labels.OnboardedFromLabel])
	require.NotContains(t, string(newRoot.HostConfig.NetworkMode), networkGroupNetworkPrefix)
	require.Equal(t, oldRootPort, requirePublishedPort(t, newRoot), "the published host port moves to the new root")

	return newRoot
}

func requireOldRootJoined(
	t *testing.T, env *TestEnvironment, group rootSwitchGroup, newRoot container.InspectResponse,
) container.InspectResponse {
	t.Helper()

	oldRoot := inspectContainer(t, env, group.oldRoot.name)

	require.NotEqual(t, group.oldRoot.id, oldRoot.ID, "the old root is recreated")
	require.True(t, oldRoot.State.Running, "the old root must run")
	require.Equal(t, container.NetworkMode(networkGroupNetworkPrefix+newRoot.ID), oldRoot.HostConfig.NetworkMode)
	require.Empty(t, oldRoot.HostConfig.PortBindings, "the old root no longer publishes ports")
	require.NotContains(t, oldRoot.Config.Labels, labels.OnboardedFromLabel)

	return oldRoot
}

func Test_RegisterContainer_RootSwitch_SingleMode_NewRootIsService(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t)

	group := startRootSwitchGroup(t, env, "rs_single", "e2e_rs_single_net")

	oldRootPort := requirePublishedPort(t, inspectContainer(t, env, group.oldRoot.name))

	registerRootSwitch(t, env, newRootSwitchRegisterRequest(group.oldRoot.id, group.serviceName, group.newRoot.id, ""))

	newRoot := requireNewRootOwnsNetwork(t, env, group, oldRootPort)
	require.Equal(t, container.NetworkMode(group.networkName), newRoot.HostConfig.NetworkMode)
	require.Equal(t, group.serviceName, newRoot.Config.Labels[labels.VervServiceLabel])
	require.Equal(t, group.serviceName, newRoot.Config.Labels[labels.DisplayNameLabel])
	require.NotContains(t, newRoot.Config.Labels, labels.Sidecar)

	aliases := newRoot.NetworkSettings.Networks[group.networkName].Aliases
	require.Contains(t, aliases, rootSwitchAlias, "the aliases of the old root move to the new root")
	require.Contains(t, aliases, group.oldRoot.name, "peers still reach the old root's name")

	oldRoot := requireOldRootJoined(t, env, group, newRoot)
	require.Equal(t, labelValueTrue, oldRoot.Config.Labels[labels.Sidecar])
	require.Equal(t, group.serviceName, oldRoot.Config.Labels[labels.VervServiceLabel])

	requireReplacedContainerKept(t, env.DockerClient(), group.oldRoot, newRoot)
	requireNoContainerNamed(t, env, group.newRoot.name+replacedContainerSuffix)

	other := requireJoinedRoot(t, env, group.other.name, group.serviceName, newRoot)
	require.NotEqual(t, group.other.id, other.ID, "single mode recreates the other member")

	requireServiceListed(t, env, group.serviceName)
	requireServiceSidecars(t, env, group.serviceName, group.oldRoot.name, group.other.name)
}

func Test_RegisterContainer_RootSwitch_SingleMode_ServiceIsAnotherMember(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t)

	group := startRootSwitchGroup(t, env, "rs_svc", "")

	oldRootPort := requirePublishedPort(t, inspectContainer(t, env, group.oldRoot.name))

	registerRootSwitch(t, env, newRootSwitchRegisterRequest(
		group.oldRoot.id, group.serviceName, group.newRoot.id, group.other.id,
	))

	newRoot := requireNewRootOwnsNetwork(t, env, group, oldRootPort)
	require.Equal(t, group.serviceName, newRoot.Config.Labels[labels.VervServiceLabel])
	require.Equal(t, labelValueTrue, newRoot.Config.Labels[labels.Sidecar])
	require.NotContains(t, newRoot.Config.Labels, labels.DisplayNameLabel)

	requireOldRootJoined(t, env, group, newRoot)

	service := inspectContainer(t, env, group.other.name)
	require.True(t, service.State.Running)
	require.Equal(t, group.serviceName, service.Config.Labels[labels.DisplayNameLabel])
	require.NotContains(t, service.Config.Labels, labels.Sidecar)
	require.Equal(t, container.NetworkMode(networkGroupNetworkPrefix+newRoot.ID), service.HostConfig.NetworkMode)

	requireServiceListed(t, env, group.serviceName)
	requireServiceSidecars(t, env, group.serviceName, group.oldRoot.name, group.newRoot.name)
}

func Test_RegisterContainer_RootSwitch_SingleMode_PreviewMatchesRegistration(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t)

	group := startRootSwitchGroup(t, env, "rs_prev", "")

	previewReq := newPreviewRegisterContainerRequest(group.oldRoot.id, group.serviceName, "")
	previewReq.RootContainerId = group.newRoot.id

	preview := previewRegisterContainer(t, env, previewReq)
	require.Len(t, preview.GetContainers(), 3)

	newRootPreview := requirePreviewOf(t, preview, group.newRoot.id)
	require.True(t, newRootPreview.GetIsRoot())
	require.True(t, newRootPreview.GetIsService())
	require.NotEmpty(t, newRootPreview.GetPorts(), "the new root previews the ports it takes over")

	oldRootPreview := requirePreviewOf(t, preview, group.oldRoot.id)
	require.False(t, oldRootPreview.GetIsRoot())
	require.Empty(t, oldRootPreview.GetPorts())

	registerRootSwitch(t, env, newRootSwitchRegisterRequest(group.oldRoot.id, group.serviceName, group.newRoot.id, ""))

	requirePreviewMatchesInspected(t, env, newRootPreview)
	requirePreviewMatchesInspected(t, env, oldRootPreview)
	requirePreviewMatchesInspected(t, env, requirePreviewOf(t, preview, group.other.id))
}

func Test_RegisterContainer_RootSwitch_ClusterMode_RecreatesRootsAndBinds(t *testing.T) {
	t.Parallel()

	env, _ := enableStatefullPgUnderDind(t, Planes[1], rootSwitchClusterSuffix)

	group := startRootSwitchGroup(t, env, "rs_cluster", "")

	oldRootPort := requirePublishedPort(t, inspectContainer(t, env, group.oldRoot.name))

	registerRootSwitch(t, env, newRootSwitchRegisterRequest(group.oldRoot.id, group.serviceName, group.newRoot.id, ""))

	newRoot := requireNewRootOwnsNetwork(t, env, group, oldRootPort)
	requireOldRootJoined(t, env, group, newRoot)

	requireJoinedRoot(t, env, group.other.name, group.serviceName, newRoot)

	boundService := env.BoundServiceName(t, group.newRoot.name, environments.DefaultEnvironmentName)
	require.Equal(t, group.serviceName, boundService)
	requireSidecarBinding(t, env, group.newRoot.name, false)
	requireSidecarBinding(t, env, group.oldRoot.name, true)
	requireSidecarBinding(t, env, group.other.name, true)

	requireServiceListed(t, env, group.serviceName)
	requireServiceSidecars(t, env, group.serviceName, group.oldRoot.name, group.other.name)
}

func Test_RegisterContainer_RootSwitch_RejectsBadRequests(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t)

	group := startRootSwitchGroup(t, env, "rs_bad", "")
	unrelated := startForeignContainer(t, env, "e2e_ng_rs_bad_unrelated", false)

	newRequest := func() *velez_api.RegisterContainer_Request {
		return newRootSwitchRegisterRequest(group.oldRoot.id, group.serviceName, group.newRoot.id, "")
	}

	t.Run("root outside the group", func(t *testing.T) {
		req := newRootSwitchRegisterRequest(group.oldRoot.id, group.serviceName, unrelated.id, "")

		requireRegisterRejected(t, env, req, user_errors.ErrRegisterRootNotInGroup)
	})

	t.Run("unknown root container", func(t *testing.T) {
		req := newRootSwitchRegisterRequest(group.oldRoot.id, group.serviceName, previewUnknownRoleId, "")

		requireRegisterRejected(t, env, req, user_errors.ErrRegisterRootNotInGroup)
	})

	t.Run("pg pattern", func(t *testing.T) {
		req := newRequest()
		req.Pattern = &velez_api.RegisterContainer_Request_Pg{Pg: &velez_api.RegisterContainer_Request_PgPattern{}}

		requireRegisterRejected(t, env, req, user_errors.ErrRegisterRootSwitchUnsupported)
	})

	t.Run("ports", func(t *testing.T) {
		req := newRequest()
		req.Ports = []*velez_api.Port{{ServicePortNumber: 80}}

		requireRegisterRejected(t, env, req, user_errors.ErrRegisterRootSwitchUnsupported)
	})

	t.Run("image tag", func(t *testing.T) {
		req := newRequest()
		req.ImageTag = toolbox.ToPtr(serviceSwitchImageTag)

		requireRegisterRejected(t, env, req, user_errors.ErrRegisterRootSwitchUnsupported)
	})

	t.Run("bind mount links", func(t *testing.T) {
		req := newRequest()
		req.BindMountLinks = []*velez_api.RegisterContainer_Request_BindMountLink{{Source: rootSwitchBindSource}}

		requireRegisterRejected(t, env, req, user_errors.ErrRegisterRootSwitchUnsupported)
	})
}

func Test_RegisterContainer_RootSwitch_NotWhitelistedOldRootChangesNothing(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t)

	rootName, newRootName, serviceName := networkGroupNames("rs_policy")

	oldRoot := startForeignContainer(t, env, rootName, true)
	newRoot := startNetworkSidecar(t, env, newRootName, oldRoot.id)

	resp := registerContainer(t, env, newRootSwitchRegisterRequest(oldRoot.id, serviceName, newRoot.id, ""))

	task := awaitRegisterTask(t, env, resp)
	require.Equal(t, velez_api.TaskStatus_FAILED, task.GetStatus())

	requireContainerUntouched(t, env.DockerClient(), oldRoot)
	requireContainerUntouched(t, env.DockerClient(), newRoot)
	requireNoContainerNamed(t, env, rootName+replacedContainerSuffix)
	requireNoContainerNamed(t, env, newRootName+replacedContainerSuffix)
}

func Test_RegisterContainer_RootSwitch_FailureRestoresTheGroup(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t)
	dockerClient := env.DockerClient()

	group := startRootSwitchGroup(t, env, "rs_roll", "")

	leftoverName := group.other.name + replacedContainerSuffix

	leftoverCfg := &container.Config{
		Image:  networkGroupSidecarImage,
		Cmd:    networkGroupSidecarCommand,
		Labels: map[string]string{testCaseNameLabel: t.Name()},
	}

	t.Cleanup(func() {
		removeForeignContainer(dockerClient, leftoverName, "")
	})

	_, err := dockerClient.ContainerCreate(t.Context(), leftoverCfg, &container.HostConfig{}, nil, nil, leftoverName)
	require.NoError(t, err)

	resp := registerContainer(t, env, newRootSwitchRegisterRequest(group.oldRoot.id, group.serviceName, group.newRoot.id, ""))

	task := awaitRegisterTask(t, env, resp)
	require.Equal(t, velez_api.TaskStatus_FAILED, task.GetStatus())

	isRestored := func() bool {
		return isContainerRestored(t, env, group.oldRoot) &&
			isContainerRestored(t, env, group.newRoot) &&
			isContainerRestored(t, env, group.other)
	}

	require.Eventually(t, isRestored, rootSwitchRollbackWait, rootSwitchRollbackTick,
		"the failed switch must put the original containers back")

	requireNoContainerNamed(t, env, group.oldRoot.name+replacedContainerSuffix)
	requireNoContainerNamed(t, env, group.newRoot.name+replacedContainerSuffix)

	_, err = dockerClient.ContainerInspect(t.Context(), leftoverName)
	require.NoError(t, err, "the container that blocked the switch is left as it was")
}

func isContainerRestored(t *testing.T, env *TestEnvironment, original foreignContainer) bool {
	t.Helper()

	inspected, err := env.DockerClient().ContainerInspect(t.Context(), original.name)
	if err != nil {
		return false
	}

	isOriginal := inspected.ID == original.id
	isUntouched := inspected.Config.Labels[labels.VervServiceLabel] == ""

	return isOriginal && isUntouched && inspected.State.Running
}
