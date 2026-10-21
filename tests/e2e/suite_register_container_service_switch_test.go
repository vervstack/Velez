//go:build e2e_full

package e2e

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"go.redsock.ru/toolbox"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service/user_errors"
	"go.vervstack.ru/Velez/internal/storage/environments"
)

const (
	serviceSwitchClusterSuffix = registerClusterSuffix + "-svcswitch"
	serviceSwitchImageTag      = "latest"
	serviceSwitchBindSource    = "/tmp/e2e_service_switch"
)

type serviceSwitchGroup struct {
	root        foreignContainer
	service     foreignContainer
	other       foreignContainer
	serviceName string
}

func startServiceSwitchGroup(
	t *testing.T, env *TestEnvironment, prefix string, isRootPublishingPort bool,
) serviceSwitchGroup {
	t.Helper()

	rootName, serviceMemberName, serviceName := networkGroupNames(prefix)
	otherName := "e2e_ng_" + prefix + "_other"

	root := startForeignContainer(t, env, rootName, isRootPublishingPort)

	return serviceSwitchGroup{
		root:        root,
		service:     startNetworkSidecar(t, env, serviceMemberName, root.id),
		other:       startNetworkSidecar(t, env, otherName, root.id),
		serviceName: serviceName,
	}
}

func newServiceSwitchRegisterRequest(rootId, serviceName, serviceContainerId string) *velez_api.RegisterContainer_Request {
	req := newRegisterContainerRequest(rootId, serviceName)
	req.ServiceContainerId = serviceContainerId

	return req
}

func newServiceSwitchPgRegisterRequest(rootId, serviceName, serviceContainerId string) *velez_api.RegisterContainer_Request {
	req := newServiceSwitchRegisterRequest(rootId, serviceName, serviceContainerId)
	req.Pattern = &velez_api.RegisterContainer_Request_Pg{
		Pg: &velez_api.RegisterContainer_Request_PgPattern{},
	}

	return req
}

func newServiceSwitchPortsRegisterRequest(rootId, serviceName, serviceContainerId string) *velez_api.RegisterContainer_Request {
	req := newServiceSwitchRegisterRequest(rootId, serviceName, serviceContainerId)
	req.Ports = []*velez_api.Port{{ServicePortNumber: 80}}

	return req
}

func newServiceSwitchImageTagRegisterRequest(
	rootId, serviceName, serviceContainerId string,
) *velez_api.RegisterContainer_Request {
	req := newServiceSwitchRegisterRequest(rootId, serviceName, serviceContainerId)
	req.ImageTag = toolbox.ToPtr(serviceSwitchImageTag)

	return req
}

func newServiceSwitchBindMountRegisterRequest(
	rootId, serviceName, serviceContainerId string,
) *velez_api.RegisterContainer_Request {
	req := newServiceSwitchRegisterRequest(rootId, serviceName, serviceContainerId)
	req.BindMountLinks = []*velez_api.RegisterContainer_Request_BindMountLink{{Source: serviceSwitchBindSource}}

	return req
}

func registerServiceSwitch(t *testing.T, env *TestEnvironment, req *velez_api.RegisterContainer_Request) {
	t.Helper()

	resp := registerContainer(t, env, req)

	task := awaitRegisterTask(t, env, resp)
	require.Equal(t, velez_api.TaskStatus_DONE, task.GetStatus(), "register task error: %s", task.GetError())
}

func requireRegisterRejected(
	t *testing.T, env *TestEnvironment, req *velez_api.RegisterContainer_Request, wantErr error,
) {
	t.Helper()

	_, err := env.VelezApiClient().RegisterContainer(t.Context(), req)
	require.Equal(t, codes.InvalidArgument, status.Code(err), "got %v", err)
	require.Contains(t, status.Convert(err).Message(), wantErr.Error())
}

func requireRecreatedRootResource(
	t *testing.T, newRoot container.InspectResponse, serviceName string,
) {
	t.Helper()

	require.True(t, newRoot.State.Running, "the new root must run")
	require.Equal(t, serviceName, newRoot.Config.Labels[labels.VervServiceLabel])
	require.Equal(t, labelValueTrue, newRoot.Config.Labels[labels.Sidecar])
	require.NotContains(t, newRoot.Config.Labels, labels.DisplayNameLabel)
	require.NotContains(t, string(newRoot.HostConfig.NetworkMode), networkGroupNetworkPrefix)
	require.Contains(t, newRoot.HostConfig.PortBindings, nat.Port("80/tcp"), "the root keeps its published port")
}

func requireRecreatedServiceMember(
	t *testing.T, service container.InspectResponse, serviceName string, newRoot container.InspectResponse,
) {
	t.Helper()

	require.True(t, service.State.Running, "the service member must run")
	require.Equal(t, serviceName, service.Config.Labels[labels.VervServiceLabel])
	require.Equal(t, serviceName, service.Config.Labels[labels.DisplayNameLabel])
	require.NotContains(t, service.Config.Labels, labels.Sidecar)
	require.NotContains(t, service.Config.Labels, labels.OnboardedFromLabel)
	require.Equal(t, container.NetworkMode(networkGroupNetworkPrefix+newRoot.ID), service.HostConfig.NetworkMode)
}

func Test_RegisterContainer_ServiceSwitch_SingleMode_RecreatesRootAndService(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t)

	group := startServiceSwitchGroup(t, env, "sw_single", true)

	registerServiceSwitch(t, env, newServiceSwitchRegisterRequest(group.root.id, group.serviceName, group.service.id))

	newRoot := inspectContainer(t, env, group.root.name)
	require.NotEqual(t, group.root.id, newRoot.ID, "single mode recreates the root")
	requireRecreatedRootResource(t, newRoot, group.serviceName)
	requireReplacedContainerKept(t, env.DockerClient(), group.root, newRoot)

	service := inspectContainer(t, env, group.service.name)
	require.NotEqual(t, group.service.id, service.ID, "single mode recreates the service member")
	requireRecreatedServiceMember(t, service, group.serviceName, newRoot)

	other := requireJoinedRoot(t, env, group.other.name, group.serviceName, newRoot)
	require.NotEqual(t, group.other.id, other.ID, "single mode recreates the other member")

	requireServiceListed(t, env, group.serviceName)
	requireServiceSidecars(t, env, group.serviceName, group.root.name, group.other.name)
}

func Test_RegisterContainer_ServiceSwitch_SingleMode_PreviewMatchesRegistration(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t)

	group := startServiceSwitchGroup(t, env, "sw_preview", false)

	previewReq := newPreviewRegisterContainerRequest(group.root.id, group.serviceName, group.service.id)
	preview := previewRegisterContainer(t, env, previewReq)
	require.Len(t, preview.GetContainers(), 3)

	registerServiceSwitch(t, env, newServiceSwitchRegisterRequest(group.root.id, group.serviceName, group.service.id))

	requirePreviewMatchesInspected(t, env, requirePreviewOf(t, preview, group.root.id))
	requirePreviewMatchesInspected(t, env, requirePreviewOf(t, preview, group.service.id))
	requirePreviewMatchesInspected(t, env, requirePreviewOf(t, preview, group.other.id))

	newRoot := inspectContainer(t, env, group.root.name)
	require.Equal(t, group.root.id, newRoot.Config.Labels[labels.OnboardedFromLabel])
}

func Test_RegisterContainer_ServiceSwitch_ClusterMode_BindsServiceMember(t *testing.T) {
	t.Parallel()

	env, _ := enableStatefullPgUnderDind(t, Planes[1], serviceSwitchClusterSuffix)

	group := startServiceSwitchGroup(t, env, "sw_cluster", false)

	registerServiceSwitch(t, env, newServiceSwitchRegisterRequest(group.root.id, group.serviceName, group.service.id))

	requireContainerUntouched(t, env.DockerClient(), group.root)
	requireContainerUntouched(t, env.DockerClient(), group.service)
	requireContainerUntouched(t, env.DockerClient(), group.other)

	require.Equal(t, group.serviceName, env.BoundServiceName(t, group.service.name, environments.DefaultEnvironmentName))
	requireSidecarBinding(t, env, group.service.name, false)
	requireSidecarBinding(t, env, group.root.name, true)
	requireSidecarBinding(t, env, group.other.name, true)

	requireServiceListed(t, env, group.serviceName)
	requireServiceSidecars(t, env, group.serviceName, group.root.name, group.other.name)
}

func Test_RegisterContainer_ServiceSwitch_RejectsBadService(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t)

	group := startServiceSwitchGroup(t, env, "sw_bad", false)
	unrelated := startForeignContainer(t, env, "e2e_ng_sw_bad_unrelated", false)

	t.Run("service outside the group", func(t *testing.T) {
		req := newServiceSwitchRegisterRequest(group.root.id, group.serviceName, unrelated.id)

		requireRegisterRejected(t, env, req, user_errors.ErrRegisterServiceNotInGroup)
	})

	t.Run("unknown service container", func(t *testing.T) {
		req := newServiceSwitchRegisterRequest(group.root.id, group.serviceName, previewUnknownRoleId)

		requireRegisterRejected(t, env, req, user_errors.ErrRegisterServiceNotInGroup)
	})

	t.Run("pg pattern", func(t *testing.T) {
		req := newServiceSwitchPgRegisterRequest(group.root.id, group.serviceName, group.service.id)

		requireRegisterRejected(t, env, req, user_errors.ErrRegisterServiceContainerUnsupported)
	})

	t.Run("ports", func(t *testing.T) {
		req := newServiceSwitchPortsRegisterRequest(group.root.id, group.serviceName, group.service.id)

		requireRegisterRejected(t, env, req, user_errors.ErrRegisterServiceContainerUnsupported)
	})

	t.Run("image tag", func(t *testing.T) {
		req := newServiceSwitchImageTagRegisterRequest(group.root.id, group.serviceName, group.service.id)

		requireRegisterRejected(t, env, req, user_errors.ErrRegisterServiceContainerUnsupported)
	})

	t.Run("bind mount links", func(t *testing.T) {
		req := newServiceSwitchBindMountRegisterRequest(group.root.id, group.serviceName, group.service.id)

		requireRegisterRejected(t, env, req, user_errors.ErrRegisterServiceContainerUnsupported)
	})
}
