//go:build e2e_full

package e2e

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"go.redsock.ru/toolbox"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/storage/environments"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

const (
	registerClusterSuffix = "e2e-register"

	registerForeignEnvKey   = "E2E_FOREIGN_ENV"
	registerForeignEnvValue = "kept"
	registerForeignMountDir = "/e2e-data"

	registerListServicesLimit  = 100
	registerDeployWatcherTicks = 12 * time.Second
)

type foreignContainer struct {
	id     string
	name   string
	volume string
}

func registerContainerNames(prefix string) (containerName, serviceName string) {
	return "e2e_reg_" + prefix + "_c", "e2e_reg_" + prefix + "_svc"
}

// startForeignContainer runs a container Velez did not create: no Velez
// labels but the per-test label the environment cleanup keys on. With
// isPublishingPort it publishes port 80 on a random host port.
func startForeignContainer(t *testing.T, env *TestEnvironment, name string, isPublishingPort bool) foreignContainer {
	t.Helper()

	dockerClient := env.Custom.NodeClients.Docker().Client()

	ctx := t.Context()
	volume := name + "_vol"

	pullReader, err := dockerClient.ImagePull(ctx, NginxAlpineImage, image.PullOptions{})
	require.NoError(t, err)

	_, err = io.Copy(io.Discard, pullReader)
	require.NoError(t, err)

	err = pullReader.Close()
	require.NoError(t, err)

	httpPort := nat.Port("80/tcp")

	cfg := &container.Config{
		Image:  NginxAlpineImage,
		Env:    []string{registerForeignEnvKey + "=" + registerForeignEnvValue},
		Labels: map[string]string{testCaseNameLabel: t.Name()},
	}

	hostCfg := &container.HostConfig{
		Binds: []string{volume + ":" + registerForeignMountDir},
	}

	if isPublishingPort {
		hostCfg.PortBindings = nat.PortMap{httpPort: []nat.PortBinding{{HostIP: "0.0.0.0"}}}
	}

	t.Cleanup(func() {
		removeForeignContainer(dockerClient, name, volume)
	})

	created, err := dockerClient.ContainerCreate(ctx, cfg, hostCfg, nil, nil, name)
	require.NoError(t, err)

	err = dockerClient.ContainerStart(ctx, created.ID, container.StartOptions{})
	require.NoError(t, err)

	return foreignContainer{id: created.ID, name: name, volume: volume}
}

func removeForeignContainer(dockerClient client.APIClient, name, volume string) {
	ctx := context.Background()

	_ = dockerClient.ContainerRemove(ctx, name, container.RemoveOptions{Force: true})
	_ = dockerClient.VolumeRemove(ctx, volume, true)
}

func newRegisterContainerRequest(containerId, serviceName string) *velez_api.RegisterContainer_Request {
	return &velez_api.RegisterContainer_Request{
		ContainerId: containerId,
		Environment: environments.DefaultEnvironmentName,
		ServiceName: serviceName,
		Pattern: &velez_api.RegisterContainer_Request_Generic{
			Generic: &velez_api.RegisterContainer_Request_GenericPattern{},
		},
	}
}

func newListContainersRequest() *velez_api.ListContainers_Request {
	return &velez_api.ListContainers_Request{
		Environment: environments.DefaultEnvironmentName,
	}
}

func newListServicesRequest() *velez_api.ListServices_Request {
	return &velez_api.ListServices_Request{
		Paging:          &velez_api.Paging{Limit: registerListServicesLimit},
		IncludeInternal: true,
	}
}

func registerContainer(
	t *testing.T,
	env *TestEnvironment,
	req *velez_api.RegisterContainer_Request,
) *velez_api.RegisterContainer_Response {
	t.Helper()

	resp, err := env.Custom.ApiGrpcImpl.RegisterContainer(t.Context(), req)
	require.NoError(t, err)
	require.NotEmpty(t, resp.GetEntityId())
	require.Equal(t, "register_container", resp.GetAction())

	return resp
}

func awaitRegisterTask(
	t *testing.T,
	env *TestEnvironment,
	resp *velez_api.RegisterContainer_Response,
) tasks_queries.VelezTask {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Minute)
	defer cancel()

	var finalTask tasks_queries.VelezTask

	for task := range env.Custom.JobsEngine.Watch(ctx, resp.GetEntityId(), resp.GetAction()) {
		finalTask = task
	}

	return finalTask
}

func findDockerContainer(
	t *testing.T,
	env *TestEnvironment,
	name string,
) *velez_api.DockerContainer {
	t.Helper()

	resp, err := env.Custom.ApiGrpcImpl.ListContainers(t.Context(), newListContainersRequest())
	require.NoError(t, err)

	for _, cont := range resp.GetContainers() {
		if cont.GetName() == name {
			return cont
		}
	}

	require.Failf(t, "container not listed", "container %q missing from ListContainers", name)

	return nil
}

func requireServiceListed(t *testing.T, env *TestEnvironment, serviceName string) {
	t.Helper()

	resp, err := env.ServiceApiClient().ListServices(t.Context(), newListServicesRequest())
	require.NoError(t, err)

	for _, svc := range resp.GetServices() {
		if svc.GetName() == serviceName {
			return
		}
	}

	require.Failf(t, "service not listed", "service %q missing from ListServices", serviceName)
}

func requireMountedVolume(t *testing.T, inspected container.InspectResponse, volume string) {
	t.Helper()

	for _, mnt := range inspected.Mounts {
		if mnt.Name == volume && mnt.Destination == registerForeignMountDir {
			return
		}
	}

	require.Failf(t, "volume mount lost", "volume %q not mounted at %q: %+v",
		volume, registerForeignMountDir, inspected.Mounts)
}

func Test_RegisterContainer_SingleMode_RecreatesWithServiceLabels(t *testing.T) {
	t.Parallel()

	containerName, serviceName := registerContainerNames("single")

	env := Planes[0].NewEnvironment(t)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	foreign := startForeignContainer(t, env, containerName, false)

	resp := registerContainer(t, env, newRegisterContainerRequest(foreign.id, serviceName))

	task := awaitRegisterTask(t, env, resp)
	require.Equal(t, tasks_queries.VelezTaskStatusDONE, task.Status, "register task error: %s", task.Error.String)

	inspected, err := dockerClient.ContainerInspect(t.Context(), containerName)
	require.NoError(t, err)
	require.NotNil(t, inspected.State)
	require.True(t, inspected.State.Running, "container must run again under the same name")
	require.NotEqual(t, foreign.id, inspected.ID, "single mode recreates the container")
	require.Equal(t, "/"+containerName, inspected.Name)

	require.Contains(t, inspected.Config.Env, registerForeignEnvKey+"="+registerForeignEnvValue)
	requireMountedVolume(t, inspected, foreign.volume)

	require.Equal(t, serviceName, inspected.Config.Labels[labels.VervServiceLabel])
	require.Equal(t, labelValueTrue, inspected.Config.Labels[labels.CreatedWithVelezLabel])

	listed := findDockerContainer(t, env, containerName)
	require.True(t, listed.GetIsRegistered())
	require.Equal(t, serviceName, listed.GetLinkedServiceName())

	requireServiceListed(t, env, serviceName)
}

func Test_RegisterContainer_ClusterMode_BindsWithoutRestart(t *testing.T) {
	containerName, serviceName := registerContainerNames("cluster")

	env, _ := enableStatefullPgUnderDind(t, Planes[1], registerClusterSuffix)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	foreign := startForeignContainer(t, env, containerName, true)

	resp := registerContainer(t, env, newRegisterContainerRequest(foreign.id, serviceName))

	task := awaitRegisterTask(t, env, resp)
	require.Equal(t, tasks_queries.VelezTaskStatusDONE, task.Status, "register task error: %s", task.Error.String)

	requireContainerUntouched(t, dockerClient, foreign)

	listed := findDockerContainer(t, env, containerName)
	require.True(t, listed.GetIsRegistered())
	require.Equal(t, serviceName, listed.GetLinkedServiceName())

	requireServiceListed(t, env, serviceName)

	bindings, err := env.Custom.Services.StorageContainer().ContainerBindings().
		ListByNode(t.Context(), domain.SelfNodeId, environments.DefaultEnvironmentName)
	require.NoError(t, err)

	var boundService string

	for _, binding := range bindings {
		if binding.ContainerName == containerName {
			boundService = binding.ServiceName
		}
	}

	require.Equal(t, serviceName, boundService, "a container binding row must exist")

	requireDeploymentRunning(t, env, serviceName)

	time.Sleep(registerDeployWatcherTicks)

	requireContainerUntouched(t, dockerClient, foreign)
	requireDeploymentRunning(t, env, serviceName)
}

func requireContainerUntouched(t *testing.T, dockerClient client.APIClient, foreign foreignContainer) {
	t.Helper()

	inspected, err := dockerClient.ContainerInspect(t.Context(), foreign.name)
	require.NoError(t, err)
	require.Equal(t, foreign.id, inspected.ID, "cluster mode must not recreate the container")
	require.True(t, inspected.State.Running)
	require.Zero(t, inspected.RestartCount)
	require.Empty(t, inspected.Config.Labels[labels.VervServiceLabel], "container labels must stay untouched")
}

func requireDeploymentRunning(t *testing.T, env *TestEnvironment, serviceName string) {
	t.Helper()

	req := &velez_api.ListDeployments_Request{ServiceName: toolbox.ToPtr(serviceName)}

	resp, err := env.Custom.ServiceApiImpl.ListDeployments(t.Context(), req)
	require.NoError(t, err)
	require.Len(t, resp.GetDeployments(), 1)
	require.Equal(t, velez_api.DeploymentStatus_RUNNING, resp.GetDeployments()[0].GetStatus())
}

func Test_RegisterContainer_UnknownContainer_TaskFails(t *testing.T) {
	t.Parallel()

	_, serviceName := registerContainerNames("unknown")

	env := Planes[0].NewEnvironment(t)

	resp := registerContainer(t, env, newRegisterContainerRequest("no-such-container-id", serviceName))

	task := awaitRegisterTask(t, env, resp)
	require.Equal(t, tasks_queries.VelezTaskStatusFAILED, task.Status)
}

func Test_RegisterContainer_AlreadyRegistered_TaskFails(t *testing.T) {
	t.Parallel()

	containerName, serviceName := registerContainerNames("twice")

	env := Planes[0].NewEnvironment(t)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	foreign := startForeignContainer(t, env, containerName, false)

	first := registerContainer(t, env, newRegisterContainerRequest(foreign.id, serviceName))
	firstTask := awaitRegisterTask(t, env, first)
	require.Equal(t, tasks_queries.VelezTaskStatusDONE, firstTask.Status, "register task error: %s", firstTask.Error.String)

	recreated, err := dockerClient.ContainerInspect(t.Context(), containerName)
	require.NoError(t, err)

	second := registerContainer(t, env, newRegisterContainerRequest(recreated.ID, serviceName))
	secondTask := awaitRegisterTask(t, env, second)
	require.Equal(t, tasks_queries.VelezTaskStatusFAILED, secondTask.Status)

	require.NotEqual(t, first.GetEntityId(), second.GetEntityId(), "every call starts a fresh task")
}

func Test_RegisterContainer_RejectsBadRequests(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t)
	velezClient := velez_api.NewVelezAPIClient(env.grpcConn)

	t.Run("empty container id", func(t *testing.T) {
		req := newRegisterContainerRequest("", "e2e_reg_bad")

		_, err := velezClient.RegisterContainer(t.Context(), req)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("empty service name", func(t *testing.T) {
		req := newRegisterContainerRequest("some-id", "")

		_, err := velezClient.RegisterContainer(t.Context(), req)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	for name, req := range newUnimplementedPatternRequests() {
		t.Run(name+" pattern is unimplemented", func(t *testing.T) {
			_, err := velezClient.RegisterContainer(t.Context(), req)
			require.Equal(t, codes.Unimplemented, status.Code(err))
		})
	}
}

func newUnimplementedPatternRequests() map[string]*velez_api.RegisterContainer_Request {
	pg := newRegisterContainerRequest("some-id", "e2e_reg_bad")
	pg.Pattern = &velez_api.RegisterContainer_Request_Pg{Pg: &velez_api.RegisterContainer_Request_PgPattern{}}

	registry := newRegisterContainerRequest("some-id", "e2e_reg_bad")
	registry.Pattern = &velez_api.RegisterContainer_Request_Registry{
		Registry: &velez_api.RegisterContainer_Request_RegistryPattern{},
	}

	runner := newRegisterContainerRequest("some-id", "e2e_reg_bad")
	runner.Pattern = &velez_api.RegisterContainer_Request_Runner{
		Runner: &velez_api.RegisterContainer_Request_RunnerPattern{},
	}

	return map[string]*velez_api.RegisterContainer_Request{"pg": pg, "registry": registry, "runner": runner}
}

func Test_RegisterContainer_EachCallStartsFreshTask(t *testing.T) {
	t.Parallel()

	_, serviceName := registerContainerNames("fresh")

	env := Planes[0].NewEnvironment(t)

	first := registerContainer(t, env, newRegisterContainerRequest("no-such-container-id", serviceName))
	firstTask := awaitRegisterTask(t, env, first)
	require.Equal(t, tasks_queries.VelezTaskStatusFAILED, firstTask.Status)

	second := registerContainer(t, env, newRegisterContainerRequest("no-such-container-id", serviceName))
	secondTask := awaitRegisterTask(t, env, second)
	require.Equal(t, tasks_queries.VelezTaskStatusFAILED, secondTask.Status)

	require.NotEqual(t, first.GetEntityId(), second.GetEntityId())
}
