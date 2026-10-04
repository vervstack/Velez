//go:build e2e_full

package e2e

import (
	"context"
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/storage/environments"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

const (
	linkTargetDir   = "/e2e-bind"
	linkHostViewDir = "/host"
	linkBeforeFile  = "before.txt"
	linkAfterFile   = "after.txt"
	linkBeforeText  = "written-before-register"
	linkAfterText   = "written-after-register"

	httpPortKey = nat.Port("80/tcp")

	replacedContainerSuffix = "_old"
)

func linkBindDir(prefix string) string {
	return "/tmp/e2e_reg_" + prefix
}

func pullNginx(t *testing.T, dockerClient client.APIClient) {
	t.Helper()

	pullReader, err := dockerClient.ImagePull(t.Context(), NginxAlpineImage, image.PullOptions{})
	require.NoError(t, err)

	_, err = io.Copy(io.Discard, pullReader)
	require.NoError(t, err)

	err = pullReader.Close()
	require.NoError(t, err)
}

// startContainerWith runs a container Velez did not create, with the given
// host config, and returns it. The container is force-removed on cleanup.
func startContainerWith(
	t *testing.T, dockerClient client.APIClient, name string, hostCfg *container.HostConfig,
) foreignContainer {
	t.Helper()

	pullNginx(t, dockerClient)

	cfg := &container.Config{
		Image:  NginxAlpineImage,
		Labels: map[string]string{testCaseNameLabel: t.Name()},
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

func startBindMountedContainer(
	t *testing.T, dockerClient client.APIClient, name, hostDir string,
) foreignContainer {
	t.Helper()

	hostCfg := &container.HostConfig{Binds: []string{hostDir + ":" + linkTargetDir}}

	return startContainerWith(t, dockerClient, name, hostCfg)
}

// startHostView runs a helper container over the same host directory, the
// only way to look at the DinD daemon's filesystem.
func startHostView(t *testing.T, dockerClient client.APIClient, name, hostDir string) foreignContainer {
	t.Helper()

	hostCfg := &container.HostConfig{Binds: []string{hostDir + ":" + linkHostViewDir}}

	return startContainerWith(t, dockerClient, name+"_hostview", hostCfg)
}

func startPortPublishingContainer(
	t *testing.T, dockerClient client.APIClient, name, hostPort string,
) foreignContainer {
	t.Helper()

	binding := nat.PortBinding{HostIP: "0.0.0.0", HostPort: hostPort}
	hostCfg := &container.HostConfig{PortBindings: nat.PortMap{httpPortKey: []nat.PortBinding{binding}}}

	return startContainerWith(t, dockerClient, name, hostCfg)
}

func execInContainer(t *testing.T, dockerClient client.APIClient, name string, cmd ...string) string {
	t.Helper()

	execOpts := container.ExecOptions{Cmd: cmd, AttachStdout: true, AttachStderr: true, Tty: true}

	created, err := dockerClient.ContainerExecCreate(t.Context(), name, execOpts)
	require.NoError(t, err)

	attachOpts := container.ExecAttachOptions{Tty: true}

	attach, err := dockerClient.ContainerExecAttach(t.Context(), created.ID, attachOpts)
	require.NoError(t, err)

	defer attach.Close()

	out, err := io.ReadAll(attach.Reader)
	require.NoError(t, err)

	inspected, err := dockerClient.ContainerExecInspect(t.Context(), created.ID)
	require.NoError(t, err)
	require.Zero(t, inspected.ExitCode, "exec %v failed: %s", cmd, out)

	return strings.TrimSpace(string(out))
}

func writeFileIn(t *testing.T, dockerClient client.APIClient, containerName, path, content string) {
	t.Helper()

	execInContainer(t, dockerClient, containerName, "sh", "-c", "echo -n "+content+" > "+path)
}

func removeVolumeOnCleanup(t *testing.T, dockerClient client.APIClient, volumeName string) {
	t.Helper()

	t.Cleanup(func() {
		_ = dockerClient.VolumeRemove(context.Background(), volumeName, true)
	})
}

func newLinkedRegisterRequest(
	containerId, serviceName, source, volumeName string,
) *velez_api.RegisterContainer_Request {
	req := newRegisterContainerRequest(containerId, serviceName)

	link := &velez_api.RegisterContainer_Request_BindMountLink{Source: source}
	if volumeName != "" {
		link.VolumeName = &volumeName
	}

	req.BindMountLinks = []*velez_api.RegisterContainer_Request_BindMountLink{link}

	return req
}

func requireRegisterDone(t *testing.T, env *TestEnvironment, req *velez_api.RegisterContainer_Request) {
	t.Helper()

	resp := registerContainer(t, env, req)

	task := awaitRegisterTask(t, env, resp)
	require.Equal(t, tasks_queries.VelezTaskStatusDONE, task.Status, "register task error: %s", task.Error.String)
}

func requireRegisterFailed(
	t *testing.T, env *TestEnvironment, req *velez_api.RegisterContainer_Request,
) tasks_queries.VelezTask {
	t.Helper()

	resp := registerContainer(t, env, req)

	task := awaitRegisterTask(t, env, resp)
	require.Equal(t, tasks_queries.VelezTaskStatusFAILED, task.Status)

	return task
}

func requireBindVolume(t *testing.T, dockerClient client.APIClient, volumeName, source string) {
	t.Helper()

	vol, err := dockerClient.VolumeInspect(t.Context(), volumeName)
	require.NoError(t, err)
	require.Equal(t, "local", vol.Driver)
	require.Equal(t, source, vol.Options["device"])
	require.Equal(t, "bind", vol.Options["o"])
	require.Equal(t, "none", vol.Options["type"])
}

func requireVolumeMountAt(t *testing.T, inspected container.InspectResponse, target, volumeName string) {
	t.Helper()

	for _, mnt := range inspected.Mounts {
		if mnt.Destination != target {
			continue
		}

		require.Equal(t, mount.TypeVolume, mnt.Type, "mount at %s: %+v", target, mnt)
		require.Equal(t, volumeName, mnt.Name)

		return
	}

	require.Failf(t, "mount missing", "no mount at %q: %+v", target, inspected.Mounts)
}

func requireSingleContainerNamed(t *testing.T, dockerClient client.APIClient, name string) {
	t.Helper()

	nameFilter := filters.NewArgs(filters.Arg("name", "^/"+name+"$"))

	listed, err := dockerClient.ContainerList(t.Context(), container.ListOptions{All: true, Filters: nameFilter})
	require.NoError(t, err)
	require.Len(t, listed, 1, "no leftover containers may share the name: %+v", listed)
}

// requireReplacedContainerKept checks the container a register replaced is
// parked as "<name>_old", not serving, and referenced from its replacement
// until onboarding is finished.
func requireReplacedContainerKept(
	t *testing.T, dockerClient client.APIClient, foreign foreignContainer, replacement container.InspectResponse,
) {
	t.Helper()

	t.Cleanup(func() {
		_ = dockerClient.ContainerRemove(context.Background(), foreign.id, container.RemoveOptions{Force: true})
	})

	replaced, err := dockerClient.ContainerInspect(t.Context(), foreign.id)
	require.NoError(t, err)
	require.Equal(t, "/"+foreign.name+replacedContainerSuffix, replaced.Name)
	require.False(t, replaced.State.Running && !replaced.State.Paused, "the replaced container must not be serving")
	require.Equal(t, foreign.id, replacement.Config.Labels[labels.OnboardedFromLabel])
}

// reservePoolHostPort takes a host port from Velez's own pool and keeps it
// locked for the test, so no parallel test is handed the same port.
func reservePoolHostPort(t *testing.T, env *TestEnvironment) uint32 {
	t.Helper()

	portManager := env.Custom.NodeClients.PortManager()

	port, err := portManager.GetPort()
	require.NoError(t, err)

	t.Cleanup(func() {
		portManager.UnlockPorts([]uint32{port})
	})

	return port
}

func requirePublishedHostPorts(t *testing.T, inspected container.InspectResponse, want ...string) {
	t.Helper()

	configured := make([]string, 0)

	for _, binding := range inspected.HostConfig.PortBindings[httpPortKey] {
		configured = append(configured, binding.HostPort)
	}

	actual := make([]string, 0)

	for _, binding := range inspected.NetworkSettings.Ports[httpPortKey] {
		actual = append(actual, binding.HostPort)
	}

	if len(want) == 0 {
		require.Empty(t, configured, "no host ports must be configured")
		require.Empty(t, actual, "no host ports must be published")

		return
	}

	require.ElementsMatch(t, want, configured)
	require.ElementsMatch(t, want, actual)
}

func actualHostPort(t *testing.T, inspected container.InspectResponse) string {
	t.Helper()

	bindings := inspected.NetworkSettings.Ports[httpPortKey]
	require.NotEmpty(t, bindings, "container must publish %s", httpPortKey)
	require.NotEmpty(t, bindings[0].HostPort)

	return bindings[0].HostPort
}

func requireServiceNotListed(t *testing.T, env *TestEnvironment, serviceName string) {
	t.Helper()

	resp, err := env.ServiceApiClient().ListServices(t.Context(), newListServicesRequest())
	require.NoError(t, err)

	for _, svc := range resp.GetServices() {
		require.NotEqual(t, serviceName, svc.GetName(), "service must not be created by a refused registration")
	}
}

func Test_RegisterContainer_BindMountNotLinked_TaskFailsUntouched(t *testing.T) {
	t.Parallel()

	containerName, serviceName := registerContainerNames("bindrefuse")
	hostDir := linkBindDir("bindrefuse")

	env := Planes[0].NewEnvironment(t)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	foreign := startBindMountedContainer(t, dockerClient, containerName, hostDir)

	task := requireRegisterFailed(t, env, newRegisterContainerRequest(foreign.id, serviceName))
	require.Contains(t, task.Error.String, hostDir)
	require.Contains(t, task.Error.String, linkTargetDir)

	inspected, err := dockerClient.ContainerInspect(t.Context(), containerName)
	require.NoError(t, err)
	require.Equal(t, foreign.id, inspected.ID)
	require.True(t, inspected.State.Running)
	require.Empty(t, inspected.Config.Labels[labels.VervServiceLabel])

	requireSingleContainerNamed(t, dockerClient, containerName)
	requireServiceNotListed(t, env, serviceName)
}

func Test_RegisterContainer_BindMountLinkUnknown_TaskFailsUntouched(t *testing.T) {
	t.Parallel()

	containerName, serviceName := registerContainerNames("linkunknown")
	hostDir := linkBindDir("linkunknown")

	env := Planes[0].NewEnvironment(t)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	foreign := startBindMountedContainer(t, dockerClient, containerName, hostDir)

	bogusSource := hostDir + "_bogus"

	req := newLinkedRegisterRequest(foreign.id, serviceName, hostDir, "")
	req.BindMountLinks = append(req.BindMountLinks, &velez_api.RegisterContainer_Request_BindMountLink{Source: bogusSource})

	task := requireRegisterFailed(t, env, req)
	require.Contains(t, task.Error.String, bogusSource)

	inspected, err := dockerClient.ContainerInspect(t.Context(), containerName)
	require.NoError(t, err)
	require.Equal(t, foreign.id, inspected.ID)

	requireServiceNotListed(t, env, serviceName)
}

func Test_RegisterContainer_SingleMode_LinksBindMountToNamedVolume(t *testing.T) {
	t.Parallel()

	containerName, serviceName := registerContainerNames("bindsingle")
	hostDir := linkBindDir("bindsingle")
	volumeName := serviceName + "_e2e-bind"

	env := Planes[0].NewEnvironment(t)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	removeVolumeOnCleanup(t, dockerClient, volumeName)

	foreign := startBindMountedContainer(t, dockerClient, containerName, hostDir)
	hostView := startHostView(t, dockerClient, containerName, hostDir)

	writeFileIn(t, dockerClient, containerName, linkTargetDir+"/"+linkBeforeFile, linkBeforeText)

	requireRegisterDone(t, env, newLinkedRegisterRequest(foreign.id, serviceName, hostDir, ""))

	inspected, err := dockerClient.ContainerInspect(t.Context(), containerName)
	require.NoError(t, err)
	require.NotEqual(t, foreign.id, inspected.ID)
	require.True(t, inspected.State.Running)
	require.Equal(t, serviceName, inspected.Config.Labels[labels.VervServiceLabel])

	requireVolumeMountAt(t, inspected, linkTargetDir, volumeName)
	requireBindVolume(t, dockerClient, volumeName, hostDir)

	got := execInContainer(t, dockerClient, containerName, "cat", linkTargetDir+"/"+linkBeforeFile)
	require.Equal(t, linkBeforeText, got)

	writeFileIn(t, dockerClient, containerName, linkTargetDir+"/"+linkAfterFile, linkAfterText)

	onHost := execInContainer(t, dockerClient, hostView.name, "cat", linkHostViewDir+"/"+linkAfterFile)
	require.Equal(t, linkAfterText, onHost)

	requireSingleContainerNamed(t, dockerClient, containerName)
}

func Test_RegisterContainer_SingleMode_LinkedBindMountKeepsExplicitVolumeName(t *testing.T) {
	t.Parallel()

	containerName, serviceName := registerContainerNames("bindnamed")
	hostDir := linkBindDir("bindnamed")
	volumeName := "e2e_reg_bindnamed_explicit"

	env := Planes[0].NewEnvironment(t)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	removeVolumeOnCleanup(t, dockerClient, volumeName)

	foreign := startBindMountedContainer(t, dockerClient, containerName, hostDir)

	requireRegisterDone(t, env, newLinkedRegisterRequest(foreign.id, serviceName, hostDir, volumeName))

	inspected, err := dockerClient.ContainerInspect(t.Context(), containerName)
	require.NoError(t, err)

	requireVolumeMountAt(t, inspected, linkTargetDir, volumeName)
	requireBindVolume(t, dockerClient, volumeName, hostDir)
}

func Test_RegisterContainer_ClusterMode_LinksAndBindings(t *testing.T) {
	env, _ := enableStatefullPgUnderDind(t, Planes[1], "e2e-register-links")
	dockerClient := env.Custom.NodeClients.Docker().Client()

	t.Run("bind mount linked keeps container and stores named volume in spec", func(t *testing.T) {
		containerName, serviceName := registerContainerNames("bindcluster")
		hostDir := linkBindDir("bindcluster")
		volumeName := "e2e_reg_bindcluster_explicit"

		removeVolumeOnCleanup(t, dockerClient, volumeName)

		foreign := startBindMountedContainer(t, dockerClient, containerName, hostDir)

		writeFileIn(t, dockerClient, containerName, linkTargetDir+"/"+linkBeforeFile, linkBeforeText)

		requireRegisterDone(t, env, newLinkedRegisterRequest(foreign.id, serviceName, hostDir, volumeName))

		requireContainerUntouched(t, dockerClient, foreign)
		requireBindVolume(t, dockerClient, volumeName, hostDir)

		got := execInContainer(t, dockerClient, containerName, "cat", linkTargetDir+"/"+linkBeforeFile)
		require.Equal(t, linkBeforeText, got)

		spec := storedSpec(t, env, serviceName)
		require.Len(t, spec.GetSettings().GetVolumes(), 1)
		require.Equal(t, volumeName, spec.GetSettings().GetVolumes()[0].GetVolumeName())
		require.Equal(t, linkTargetDir, spec.GetSettings().GetVolumes()[0].GetContainerPath())
	})

	t.Run("random host port is listed with its actual value", func(t *testing.T) {
		containerName, serviceName := registerContainerNames("randcluster")

		foreign := startPortPublishingContainer(t, dockerClient, containerName, "")

		inspectedBefore, err := dockerClient.ContainerInspect(t.Context(), containerName)
		require.NoError(t, err)

		actualPort := actualHostPort(t, inspectedBefore)

		requireRegisterDone(t, env, newRegisterContainerRequest(foreign.id, serviceName))

		listed := findDockerContainer(t, env, containerName)
		require.True(t, listed.GetIsRegistered())

		var exposed []string

		for _, port := range listed.GetPorts() {
			if port.GetExposedTo() != 0 {
				exposed = append(exposed, strconv.Itoa(int(port.GetExposedTo())))
			}
		}

		require.Contains(t, exposed, actualPort)

		spec := storedSpec(t, env, serviceName)
		require.Len(t, spec.GetSettings().GetPorts(), 1)
		require.Equal(t, actualPort, strconv.Itoa(int(spec.GetSettings().GetPorts()[0].GetExposedTo())))
	})

	t.Run("empty environment resolves like the default one", func(t *testing.T) {
		containerName, serviceName := registerContainerNames("envempty")

		foreign := startForeignContainer(t, env, containerName, false)

		req := newRegisterContainerRequest(foreign.id, serviceName)
		req.Environment = ""

		requireRegisterDone(t, env, req)

		byDefault := findDockerContainer(t, env, containerName)
		require.Equal(t, serviceName, byDefault.GetLinkedServiceName(), "ListContainers(PROD) must see the binding")

		emptyReq := &velez_api.ListContainers_Request{}

		resp, err := env.Custom.ApiGrpcImpl.ListContainers(t.Context(), emptyReq)
		require.NoError(t, err)

		var linkedByEmpty string

		for _, cont := range resp.GetContainers() {
			if cont.GetName() == containerName {
				linkedByEmpty = cont.GetLinkedServiceName()
			}
		}

		require.Equal(t, serviceName, linkedByEmpty, "ListContainers(\"\") must see the binding")

		bindings, err := env.Custom.Services.StorageContainer().ContainerBindings().
			ListByNode(t.Context(), domain.SelfNodeId, environments.DefaultEnvironmentName)
		require.NoError(t, err)

		var boundService string

		for _, binding := range bindings {
			if binding.ContainerName == containerName {
				boundService = binding.ServiceName
			}
		}

		require.Equal(t, serviceName, boundService, "binding must be stored under the default environment name")
	})
}

func storedSpec(t *testing.T, env *TestEnvironment, serviceName string) *velez_api.CreateSmerd_Request {
	t.Helper()

	deployments := env.Custom.Services.StorageContainer().Deployments()

	list, err := deployments.List(t.Context(), domain.ListDeploymentsReq{ServiceName: serviceName})
	require.NoError(t, err)
	require.Len(t, list, 1)

	row, err := deployments.GetSpecificationById(t.Context(), list[0].SpecId)
	require.NoError(t, err)
	require.True(t, row.VervPayload.Valid)

	spec := &velez_api.CreateSmerd_Request{}

	err = json.Unmarshal(row.VervPayload.RawMessage, spec)
	require.NoError(t, err)

	return spec
}

func newPortRegisterRequest(containerId, serviceName string, isKeepMapping bool, hostPorts ...uint32) *velez_api.RegisterContainer_Request {
	req := newRegisterContainerRequest(containerId, serviceName)
	req.KeepPortMapping = isKeepMapping

	for _, hostPort := range hostPorts {
		port := &velez_api.Port{
			ServicePortNumber: 80,
			Protocol:          velez_api.Port_tcp,
			ExposedTo:         &hostPort,
		}

		req.Ports = append(req.Ports, port)
	}

	return req
}

func Test_RegisterContainer_SingleMode_Ports(t *testing.T) {
	t.Run("keep mapping recreates with the same host port", func(t *testing.T) {
		t.Parallel()

		containerName, serviceName := registerContainerNames("portkeep")
		env := Planes[0].NewEnvironment(t)
		dockerClient := env.Custom.NodeClients.Docker().Client()

		foreign := startPortPublishingContainer(t, dockerClient, containerName, "31101")

		requireRegisterDone(t, env, newPortRegisterRequest(foreign.id, serviceName, true))

		inspected, err := dockerClient.ContainerInspect(t.Context(), containerName)
		require.NoError(t, err)
		require.NotEqual(t, foreign.id, inspected.ID)
		require.True(t, inspected.State.Running)
		require.Equal(t, serviceName, inspected.Config.Labels[labels.VervServiceLabel])

		requirePublishedHostPorts(t, inspected, "31101")
		requireReplacedContainerKept(t, dockerClient, foreign, inspected)
		requireSingleContainerNamed(t, dockerClient, containerName)
	})

	t.Run("explicit ports replace the mapping", func(t *testing.T) {
		t.Parallel()

		containerName, serviceName := registerContainerNames("portswap")
		env := Planes[0].NewEnvironment(t)
		dockerClient := env.Custom.NodeClients.Docker().Client()

		foreign := startPortPublishingContainer(t, dockerClient, containerName, "31102")

		targetHostPort := reservePoolHostPort(t, env)
		isHeld := env.Custom.NodeClients.PortManager().HoldPort(targetHostPort)
		require.True(t, isHeld)

		requireRegisterDone(t, env, newPortRegisterRequest(foreign.id, serviceName, false, targetHostPort))

		inspected, err := dockerClient.ContainerInspect(t.Context(), containerName)
		require.NoError(t, err)
		require.True(t, inspected.State.Running)

		requirePublishedHostPorts(t, inspected, strconv.FormatUint(uint64(targetHostPort), 10))
		requireReplacedContainerKept(t, dockerClient, foreign, inspected)
		requireSingleContainerNamed(t, dockerClient, containerName)
	})

	t.Run("no ports publishes nothing", func(t *testing.T) {
		t.Parallel()

		containerName, serviceName := registerContainerNames("portnone")
		env := Planes[0].NewEnvironment(t)
		dockerClient := env.Custom.NodeClients.Docker().Client()

		foreign := startPortPublishingContainer(t, dockerClient, containerName, "31104")

		requireRegisterDone(t, env, newPortRegisterRequest(foreign.id, serviceName, false))

		inspected, err := dockerClient.ContainerInspect(t.Context(), containerName)
		require.NoError(t, err)
		require.True(t, inspected.State.Running)

		requirePublishedHostPorts(t, inspected)
		requireReplacedContainerKept(t, dockerClient, foreign, inspected)
		requireSingleContainerNamed(t, dockerClient, containerName)
	})

	t.Run("port conflict rolls the old container back", func(t *testing.T) {
		t.Parallel()

		containerName, serviceName := registerContainerNames("portrollback")
		env := Planes[0].NewEnvironment(t)
		dockerClient := env.Custom.NodeClients.Docker().Client()

		foreign := startPortPublishingContainer(t, dockerClient, containerName, "31105")

		lockedHostPort := reservePoolHostPort(t, env)

		requireRegisterFailed(t, env, newPortRegisterRequest(foreign.id, serviceName, false, lockedHostPort))

		inspected, err := dockerClient.ContainerInspect(t.Context(), containerName)
		require.NoError(t, err)
		require.Equal(t, foreign.id, inspected.ID, "the original container must be restored")
		require.True(t, inspected.State.Running)
		require.False(t, inspected.State.Paused)
		require.Empty(t, inspected.Config.Labels[labels.VervServiceLabel])

		requirePublishedHostPorts(t, inspected, "31105")
		requireSingleContainerNamed(t, dockerClient, containerName)
	})

	t.Run("random host port is kept by keep mapping", func(t *testing.T) {
		t.Parallel()

		containerName, serviceName := registerContainerNames("portrandom")
		env := Planes[0].NewEnvironment(t)
		dockerClient := env.Custom.NodeClients.Docker().Client()

		foreign := startPortPublishingContainer(t, dockerClient, containerName, "")

		before, err := dockerClient.ContainerInspect(t.Context(), containerName)
		require.NoError(t, err)

		randomPort := actualHostPort(t, before)

		requireRegisterDone(t, env, newPortRegisterRequest(foreign.id, serviceName, true))

		inspected, err := dockerClient.ContainerInspect(t.Context(), containerName)
		require.NoError(t, err)
		require.NotEqual(t, foreign.id, inspected.ID)
		require.True(t, inspected.State.Running)

		requirePublishedHostPorts(t, inspected, randomPort)
	})
}

func Test_RegisterContainer_RejectsPortsWithKeepMapping(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t)
	velezClient := velez_api.NewVelezAPIClient(env.grpcConn)

	req := newPortRegisterRequest("some-id", "e2e_reg_keepports", true, 31106)

	_, err := velezClient.RegisterContainer(t.Context(), req)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}
