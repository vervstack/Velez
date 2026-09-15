//go:build e2e_full

package e2e

import (
	"context"
	"sync"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

const (
	// containerRegistryLifecycleInstanceName is a fixed, suite-unique name -
	// unlike the retired registry plugin's node-singleton container name,
	// CreateRegistryInstance.Request.Name is caller-chosen and this one
	// doesn't collide with any other suite's fixed names, so this suite is
	// safe to run t.Parallel() against its siblings; every test
	// force-removes its own container(s)/volumes before and after
	// regardless.
	containerRegistryLifecycleInstanceName = "e2e-container-registry-lifecycle"

	containerRegistryCollisionInstanceNameA = "e2e-container-registry-collision-a"
	containerRegistryCollisionInstanceNameB = "e2e-container-registry-collision-b"
)

// ContainerRegistryLifecycleSuite exercises Container-Registry-as-a-Service
// end to end in plain single-node/local_storage mode (no WithMatreshka) -
// mirrors PgaasLifecycleSuite, the sibling *aas feature this one was built
// from. RegistryaasService.CreateRegistryInstance itself blocks on the
// create_registry_instance task (see
// internal/service/service_manager/registryaas/create.go), but that task
// only gets as far as writing SCHEDULED_DEPLOYMENT rows for the registry and
// its UI sidecar (deployRegistryInstanceJob/deployRegistryUiJob hand off to
// CreateNewDeploy - see its doc comment); the deploy watcher
// (internal/workers/deploy_watcher.go) is what actually dispatches and runs
// create_smerd for each. So this suite additionally watches both create_smerd
// tasks before asserting the instance is up.
type ContainerRegistryLifecycleSuite struct {
	suite.Suite

	plane Plane
}

func (s *ContainerRegistryLifecycleSuite) Test_ContainerRegistryLifecycle_HappyPath() {
	t := s.T()
	t.Parallel()

	ctx := t.Context()

	env := s.plane.NewEnvironment(t)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	removeContainerRegistryInstance(dockerClient, containerRegistryLifecycleInstanceName)
	t.Cleanup(func() { removeContainerRegistryInstance(dockerClient, containerRegistryLifecycleInstanceName) })

	createReq := &velez_api.CreateRegistryInstance_Request{Name: containerRegistryLifecycleInstanceName}

	_, err := env.Custom.ContainerRegistryApiImpl.CreateRegistryInstance(ctx, createReq)
	require.NoError(t, err)

	waitForRegistryInstanceDeploy(t, env, containerRegistryLifecycleInstanceName)

	instance := findRegistryInstance(t, env, containerRegistryLifecycleInstanceName)
	require.NotNil(t, instance, "expected registry instance %q in ListRegistryInstances",
		containerRegistryLifecycleInstanceName)
	require.Equal(t, containerRegistryLifecycleInstanceName, instance.GetName())
	require.NotZero(t, instance.GetPort())
	require.NotZero(t, instance.GetUiPort())
	require.NotEmpty(t, instance.GetUsername())

	credsReq := &velez_api.GetRegistryInstanceCredentials_Request{Name: containerRegistryLifecycleInstanceName}

	credsResp, err := env.Custom.ContainerRegistryApiImpl.GetRegistryInstanceCredentials(ctx, credsReq)
	require.NoError(t, err)
	require.Equal(t, instance.GetUsername(), credsResp.GetUsername())
	require.NotEmpty(t, credsResp.GetPassword())
	require.NotEmpty(t, credsResp.GetRegistryUrl())

	dropReq := &velez_api.DropRegistryInstance_Request{Name: containerRegistryLifecycleInstanceName}

	_, err = env.Custom.ContainerRegistryApiImpl.DropRegistryInstance(ctx, dropReq)
	require.NoError(t, err)

	dropped := findRegistryInstance(t, env, containerRegistryLifecycleInstanceName)
	require.Nil(t, dropped, "registry instance still listed after drop")
}

// Test_ContainerRegistryLifecycle_TwoInstances_NoCollision creates two
// registry instances concurrently and asserts neither's port, container, or
// volume allocation clobbers the other's - resolveRegistryPortsJob draws
// from the node's shared PortManager pool for both instances' registry and
// UI sidecar ports, and buildRegistryDeployRequest derives per-instance
// volume names from each instance's own name, so nothing here is expected to
// collide; this proves it under real concurrent creation rather than by
// inspection.
func (s *ContainerRegistryLifecycleSuite) Test_ContainerRegistryLifecycle_TwoInstances_NoCollision() {
	t := s.T()
	t.Parallel()

	ctx := t.Context()

	env := s.plane.NewEnvironment(t)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	names := []string{containerRegistryCollisionInstanceNameA, containerRegistryCollisionInstanceNameB}
	for _, name := range names {
		removeContainerRegistryInstance(dockerClient, name)
		t.Cleanup(func(name string) func() { return func() { removeContainerRegistryInstance(dockerClient, name) } }(name))
	}

	var wg sync.WaitGroup

	errs := make([]error, len(names))

	for i, name := range names {
		wg.Add(1)

		go func(i int, name string) {
			defer wg.Done()

			createReq := &velez_api.CreateRegistryInstance_Request{Name: name}
			_, createErr := env.Custom.ContainerRegistryApiImpl.CreateRegistryInstance(ctx, createReq)

			errs[i] = createErr
		}(i, name)
	}

	wg.Wait()

	for i, name := range names {
		require.NoError(t, errs[i], "error creating registry instance %q", name)
	}

	for _, name := range names {
		waitForRegistryInstanceDeploy(t, env, name)
	}

	instanceA := findRegistryInstance(t, env, containerRegistryCollisionInstanceNameA)
	instanceB := findRegistryInstance(t, env, containerRegistryCollisionInstanceNameB)
	require.NotNil(t, instanceA)
	require.NotNil(t, instanceB)

	require.NotEqual(t, instanceA.GetPort(), instanceB.GetPort(), "registry instances share a container port")
	require.NotEqual(t, instanceA.GetUiPort(), instanceB.GetUiPort(), "registry instances share a ui port")

	inspectA, err := dockerClient.ContainerInspect(ctx, containerRegistryCollisionInstanceNameA)
	require.NoError(t, err)
	require.True(t, inspectA.State.Running)

	inspectB, err := dockerClient.ContainerInspect(ctx, containerRegistryCollisionInstanceNameB)
	require.NoError(t, err)
	require.True(t, inspectB.State.Running)

	inspectUiA, err := dockerClient.ContainerInspect(ctx, containerRegistryCollisionInstanceNameA+"-ui")
	require.NoError(t, err)
	require.True(t, inspectUiA.State.Running)

	inspectUiB, err := dockerClient.ContainerInspect(ctx, containerRegistryCollisionInstanceNameB+"-ui")
	require.NoError(t, err)
	require.True(t, inspectUiB.State.Running)
}

func Test_ContainerRegistryLifecycle(t *testing.T) {
	t.Parallel()
	RunPlaneSuite(t, Planes, func(plane Plane) suite.TestingSuite {
		return &ContainerRegistryLifecycleSuite{plane: plane}
	})
}

// findRegistryInstance returns the listed registry instance with the given
// name, or nil.
func findRegistryInstance(t *testing.T, env *TestEnvironment, name string) *velez_api.RegistryInstance {
	t.Helper()

	listReq := &velez_api.ListRegistryInstances_Request{}

	listResp, err := env.Custom.ContainerRegistryApiImpl.ListRegistryInstances(t.Context(), listReq)
	require.NoError(t, err)

	for _, instance := range listResp.GetInstances() {
		if instance.GetName() == name {
			return instance
		}
	}

	return nil
}

// waitForRegistryInstanceDeploy watches the create_smerd tasks the deploy
// watcher dispatches for a registry instance and its UI sidecar - see
// ContainerRegistryLifecycleSuite's doc comment on why CreateRegistryInstance
// returning isn't itself proof either container exists yet.
func waitForRegistryInstanceDeploy(t *testing.T, env *TestEnvironment, instanceName string) {
	t.Helper()

	ctx := t.Context()

	registryEntityID := jobs.SmerdEntityID("", instanceName)

	var registryTask tasks_queries.VelezTask

	for task := range env.Custom.JobsEngine.Watch(ctx, registryEntityID, jobs.CreateSmerdAction) {
		registryTask = task
	}

	require.Equal(t, tasks_queries.VelezTaskStatusDONE, registryTask.Status,
		"create_smerd task for registry instance %q error: %s", instanceName, registryTask.Error.String)

	uiEntityID := jobs.SmerdEntityID("", instanceName+"-ui")

	var uiTask tasks_queries.VelezTask

	for task := range env.Custom.JobsEngine.Watch(ctx, uiEntityID, jobs.CreateSmerdAction) {
		uiTask = task
	}

	require.Equal(t, tasks_queries.VelezTaskStatusDONE, uiTask.Status,
		"create_smerd task for registry ui sidecar %q error: %s", instanceName+"-ui", uiTask.Error.String)
}

// removeContainerRegistryInstance force-removes an instance's registry
// container, its UI sidecar container, and its per-instance data/auth
// volumes (registryaas.pgVolumeName-style "<name>-data"/"<name>-auth"),
// ignoring "no such container/volume" - mirrors removePgaasInstance.
func removeContainerRegistryInstance(dockerClient client.APIClient, name string) {
	ctx := context.Background()
	removeOpts := container.RemoveOptions{Force: true}

	_ = dockerClient.ContainerRemove(ctx, name, removeOpts)
	_ = dockerClient.ContainerRemove(ctx, name+"-ui", removeOpts)
	_ = dockerClient.VolumeRemove(ctx, name+"-data", true)
	_ = dockerClient.VolumeRemove(ctx, name+"-auth", true)
}
