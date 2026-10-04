//go:build e2e_full

package e2e

import (
	"context"
	"io"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/storage/environments"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	networkGroupClusterSuffix = registerClusterSuffix + "-netgroup"
	networkGroupNetworkPrefix = "container:"

	// A sidecar joining the root's network namespace is an elevated request, so
	// the runtime policy only lets a whitelisted image (the tailscale sidecar
	// in production) do it; nginx is tagged under that name as a stand-in.
	networkGroupSidecarImage = "tailscale/tailscale:e2e"
)

var networkGroupSidecarCommand = []string{"sleep", "3600"}

func networkGroupNames(prefix string) (rootName, sidecarName, serviceName string) {
	return "e2e_ng_" + prefix + "_root", "e2e_ng_" + prefix + "_side", "e2e_ng_" + prefix + "_svc"
}

func tagNginxAsSidecarImage(t *testing.T, env *TestEnvironment) {
	t.Helper()

	dockerClient := env.Custom.NodeClients.Docker().Client()

	pullReader, err := dockerClient.ImagePull(t.Context(), NginxAlpineImage, image.PullOptions{})
	require.NoError(t, err)

	_, err = io.Copy(io.Discard, pullReader)
	require.NoError(t, err)

	err = pullReader.Close()
	require.NoError(t, err)

	err = dockerClient.ImageTag(t.Context(), NginxAlpineImage, networkGroupSidecarImage)
	require.NoError(t, err)
}

func startNetworkSidecar(
	t *testing.T, env *TestEnvironment, name, rootId string,
) foreignContainer {
	t.Helper()

	dockerClient := env.Custom.NodeClients.Docker().Client()

	tagNginxAsSidecarImage(t, env)

	cfg := &container.Config{
		Image:  networkGroupSidecarImage,
		Cmd:    networkGroupSidecarCommand,
		Labels: map[string]string{testCaseNameLabel: t.Name()},
	}

	hostCfg := &container.HostConfig{
		NetworkMode: container.NetworkMode(networkGroupNetworkPrefix + rootId),
	}

	t.Cleanup(func() {
		removeForeignContainer(dockerClient, name, "")
	})

	created, err := dockerClient.ContainerCreate(t.Context(), cfg, hostCfg, nil, nil, name)
	require.NoError(t, err)

	err = dockerClient.ContainerStart(t.Context(), created.ID, container.StartOptions{})
	require.NoError(t, err)

	return foreignContainer{id: created.ID, name: name}
}

func newGetServiceRequest(serviceName string) *velez_api.GetService_Request {
	return &velez_api.GetService_Request{Name: serviceName}
}

func newNetworkGroupListSmerdsRequest(t *testing.T) *velez_api.ListSmerds_Request {
	t.Helper()

	return &velez_api.ListSmerds_Request{
		Label: map[string]string{testCaseNameLabel: t.Name()},
	}
}

func newNetworkGroupUpgradeRequest(rootName string) *velez_api.UpgradeSmerd_Request {
	return &velez_api.UpgradeSmerd_Request{
		Name:  rootName,
		Image: NginxAlpineImage,
	}
}

func registerNetworkGroupRoot(
	t *testing.T, env *TestEnvironment, rootId, serviceName string,
) {
	t.Helper()

	resp := registerContainer(t, env, newRegisterContainerRequest(rootId, serviceName))

	task := awaitRegisterTask(t, env, resp)
	require.Equal(t, tasks_queries.VelezTaskStatusDONE, task.Status, "register task error: %s", task.Error.String)
}

func inspectContainer(t *testing.T, env *TestEnvironment, name string) container.InspectResponse {
	t.Helper()

	dockerClient := env.Custom.NodeClients.Docker().Client()

	inspected, err := dockerClient.ContainerInspect(t.Context(), name)
	require.NoError(t, err)

	return inspected
}

func requireJoinedRoot(
	t *testing.T, env *TestEnvironment, sidecarName, serviceName string, root container.InspectResponse,
) container.InspectResponse {
	t.Helper()

	sidecar := inspectContainer(t, env, sidecarName)

	require.True(t, sidecar.State.Running, "sidecar must run")
	require.Equal(t, labelValueTrue, sidecar.Config.Labels[labels.Sidecar])
	require.Equal(t, serviceName, sidecar.Config.Labels[labels.VervServiceLabel])
	require.Equal(t, container.NetworkMode(networkGroupNetworkPrefix+root.ID), sidecar.HostConfig.NetworkMode)

	return sidecar
}

func requireNotListedAsSmerd(t *testing.T, env *TestEnvironment, containerName string) {
	t.Helper()

	resp := env.ListSmerds(t, t.Context(), newNetworkGroupListSmerdsRequest(t))

	for _, smerd := range resp.GetSmerds() {
		require.NotEqual(t, containerName, smerd.GetName(), "a sidecar must not be listed as a smerd")
	}
}

func requireServiceSidecars(t *testing.T, env *TestEnvironment, serviceName string, want ...string) {
	t.Helper()

	resp, err := env.ServiceApiClient().GetService(t.Context(), newGetServiceRequest(serviceName))
	require.NoError(t, err)

	got := make([]string, 0, len(resp.GetSidecars()))

	for _, sidecar := range resp.GetSidecars() {
		require.NotEmpty(t, sidecar.GetContainerId())
		require.NotEmpty(t, sidecar.GetImageName())
		require.Equal(t, velez_api.Smerd_running, sidecar.GetStatus())

		got = append(got, sidecar.GetContainerName())
	}

	require.ElementsMatch(t, want, got)
}

func requireSidecarBinding(t *testing.T, env *TestEnvironment, containerName string, isSidecar bool) {
	t.Helper()

	bindings, err := env.Custom.Services.StorageContainer().ContainerBindings().
		ListByNode(t.Context(), domain.SelfNodeId, environments.DefaultEnvironmentName)
	require.NoError(t, err)

	for _, binding := range bindings {
		if binding.ContainerName == containerName {
			require.Equal(t, isSidecar, binding.IsSidecar)

			return
		}
	}

	require.Failf(t, "binding missing", "no container binding for %q", containerName)
}

func Test_NetworkGroup_ListMarksSidecarOwner(t *testing.T) {
	t.Parallel()

	rootName, sidecarName, _ := networkGroupNames("list")

	env := Planes[0].NewEnvironment(t)

	root := startForeignContainer(t, env, rootName, false)
	startNetworkSidecar(t, env, sidecarName, root.id)

	listedSidecar := findDockerContainer(t, env, sidecarName)
	require.Equal(t, root.id, listedSidecar.GetNetworkOwnerContainerId())

	listedRoot := findDockerContainer(t, env, rootName)
	require.Nil(t, listedRoot.NetworkOwnerContainerId)
}

func Test_NetworkGroup_RegisterSingleMode(t *testing.T) {
	t.Parallel()

	rootName, sidecarName, serviceName := networkGroupNames("single")

	env := Planes[0].NewEnvironment(t)

	root := startForeignContainer(t, env, rootName, false)
	oldSidecar := startNetworkSidecar(t, env, sidecarName, root.id)

	registerNetworkGroupRoot(t, env, root.id, serviceName)

	newRoot := inspectContainer(t, env, rootName)
	require.NotEqual(t, root.id, newRoot.ID, "single mode recreates the root")
	require.Equal(t, serviceName, newRoot.Config.Labels[labels.VervServiceLabel])

	newSidecar := requireJoinedRoot(t, env, sidecarName, serviceName, newRoot)
	require.NotEqual(t, oldSidecar.id, newSidecar.ID, "single mode recreates the sidecar")

	requireNotListedAsSmerd(t, env, sidecarName)
	requireServiceSidecars(t, env, serviceName, sidecarName)
}

func Test_NetworkGroup_RegisterClusterMode(t *testing.T) {
	rootName, sidecarName, serviceName := networkGroupNames("cluster")

	env, _ := enableStatefullPgUnderDind(t, Planes[1], networkGroupClusterSuffix)

	root := startForeignContainer(t, env, rootName, false)
	sidecar := startNetworkSidecar(t, env, sidecarName, root.id)

	registerNetworkGroupRoot(t, env, root.id, serviceName)

	require.Equal(t, root.id, inspectContainer(t, env, rootName).ID, "cluster mode keeps the root")
	require.Equal(t, sidecar.id, inspectContainer(t, env, sidecarName).ID, "cluster mode keeps the sidecar")

	requireSidecarBinding(t, env, rootName, false)
	requireSidecarBinding(t, env, sidecarName, true)

	requireServiceSidecars(t, env, serviceName, sidecarName)
}

func Test_NetworkGroup_RegisterSidecarIsRefused(t *testing.T) {
	t.Parallel()

	rootName, sidecarName, serviceName := networkGroupNames("refuse")

	env := Planes[0].NewEnvironment(t)

	root := startForeignContainer(t, env, rootName, false)
	sidecar := startNetworkSidecar(t, env, sidecarName, root.id)

	_, err := env.Custom.ApiGrpcImpl.RegisterContainer(t.Context(), newRegisterContainerRequest(sidecar.id, serviceName))
	require.True(t, rerrors.Is(err, user_errors.ErrContainerSharesNetwork), "got %v", err)
}

func Test_NetworkGroup_PartialGroupAdoptsNewSidecar(t *testing.T) {
	t.Parallel()

	rootName, sidecarName, serviceName := networkGroupNames("partial")

	env := Planes[0].NewEnvironment(t)

	root := startForeignContainer(t, env, rootName, false)

	registerNetworkGroupRoot(t, env, root.id, serviceName)

	registeredRoot := inspectContainer(t, env, rootName)

	startNetworkSidecar(t, env, sidecarName, registeredRoot.ID)

	registerNetworkGroupRoot(t, env, registeredRoot.ID, serviceName)

	require.Equal(t, registeredRoot.ID, inspectContainer(t, env, rootName).ID, "a registered root is not recreated")

	requireJoinedRoot(t, env, sidecarName, serviceName, registeredRoot)
	requireServiceSidecars(t, env, serviceName, sidecarName)
}

func Test_NetworkGroup_UpgradeReattachesSidecar(t *testing.T) {
	t.Parallel()

	rootName, sidecarName, serviceName := networkGroupNames("upgrade")

	env := Planes[0].NewEnvironment(t)

	root := startForeignContainer(t, env, rootName, false)
	startNetworkSidecar(t, env, sidecarName, root.id)

	registerNetworkGroupRoot(t, env, root.id, serviceName)

	registeredRoot := inspectContainer(t, env, rootName)
	registeredSidecar := inspectContainer(t, env, sidecarName)

	upgrade(t, env.Custom.ApiGrpcImpl, newNetworkGroupUpgradeRequest(rootName))

	upgradedRoot := inspectContainer(t, env, rootName)
	require.NotEqual(t, registeredRoot.ID, upgradedRoot.ID, "upgrade replaces the root")

	upgradedSidecar := requireJoinedRoot(t, env, sidecarName, serviceName, upgradedRoot)
	require.NotEqual(t, registeredSidecar.ID, upgradedSidecar.ID, "upgrade recreates the sidecar")

	requireServiceSidecars(t, env, serviceName, sidecarName)
}

type upgradeCaller interface {
	UpgradeSmerd(context.Context, *velez_api.UpgradeSmerd_Request) (*velez_api.UpgradeSmerd_Response, error)
}

func upgrade(t *testing.T, caller upgradeCaller, req *velez_api.UpgradeSmerd_Request) {
	t.Helper()

	_, err := caller.UpgradeSmerd(t.Context(), req)
	require.NoError(t, err)
}
