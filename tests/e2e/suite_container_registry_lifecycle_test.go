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
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/jobs"
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

	containerRegistryCollisionInstanceNameA = "e2e-cr-collision-a"
	containerRegistryCollisionInstanceNameB = "e2e-cr-collision-b"
)

// ContainerRegistryLifecycleSuite exercises Container-Registry-as-a-Service
// end to end in plain single-node/local_storage mode (no WithMatreshka) -
// mirrors PgaasLifecycleSuite, the sibling *aas feature this one was built
// from. RegistryaasService.CreateRegistryInstance only enqueues the
// create_registry_instance task and returns (see
// internal/service/service_manager/registryaas/create.go); that task itself
// only gets as far as writing SCHEDULED_DEPLOYMENT rows for the registry and
// its optional UI sidecar (deployRegistryInstanceJob/deployRegistryUiJob hand
// off to CreateNewDeploy - see its doc comment). The deploy watcher
// (internal/workers/deploy_watcher.go) is what actually dispatches and runs
// create_smerd for each. So this suite watches the create_registry_instance
// task itself (proves the instance/registry_instances rows are written) and
// then the create_smerd task(s) the deploy watcher dispatches (proves the
// container(s) are actually up) before asserting anything.
type ContainerRegistryLifecycleSuite struct {
	suite.Suite

	plane Plane
}

// registryServiceName is the name registryaas gives the service, containers,
// volumes and create_smerd task entities: the caller's name behind
// labels.RegistryaasNamePrefix.
func registryServiceName(instanceName string) string {
	return labels.RegistryaasNamePrefix + instanceName
}

func (s *ContainerRegistryLifecycleSuite) Test_ContainerRegistryLifecycle_HappyPath() {
	t := s.T()
	t.Parallel()

	ctx := t.Context()

	env := s.plane.NewEnvironment(t)
	dockerClient := env.DockerClient()

	removeContainerRegistryInstance(dockerClient, containerRegistryLifecycleInstanceName)
	t.Cleanup(func() { removeContainerRegistryInstance(dockerClient, containerRegistryLifecycleInstanceName) })

	createReq := newCreateRegistryInstanceRequest(containerRegistryLifecycleInstanceName, false)

	_, err := env.ContainerRegistryClient().CreateRegistryInstance(ctx, createReq)
	require.NoError(t, err)

	waitForRegistryInstanceDeploy(t, env, containerRegistryLifecycleInstanceName, false)

	instance := findRegistryInstance(t, env, registryServiceName(containerRegistryLifecycleInstanceName))
	require.NotNil(t, instance, "expected registry instance %q in ListRegistryInstances",
		containerRegistryLifecycleInstanceName)
	require.Equal(t, registryServiceName(containerRegistryLifecycleInstanceName), instance.GetName())
	require.NotZero(t, instance.GetPort())
	require.Zero(t, instance.GetUiPort(), "ui sidecar should not be provisioned when enable_ui is unset")
	require.NotEmpty(t, instance.GetUsername())

	credsReq := &velez_api.GetRegistryInstanceCredentials_Request{Name: registryServiceName(containerRegistryLifecycleInstanceName)}

	credsResp, err := env.ContainerRegistryClient().GetRegistryInstanceCredentials(ctx, credsReq)
	require.NoError(t, err)
	require.Equal(t, instance.GetUsername(), credsResp.GetUsername())
	require.NotEmpty(t, credsResp.GetPassword())
	require.NotEmpty(t, credsResp.GetRegistryUrl())

	dropRegistryInstanceAndAwait(t, env, registryServiceName(containerRegistryLifecycleInstanceName))

	dropped := findRegistryInstance(t, env, registryServiceName(containerRegistryLifecycleInstanceName))
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
	dockerClient := env.DockerClient()

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

			createReq := newCreateRegistryInstanceRequest(name, true)
			_, createErr := env.ContainerRegistryClient().CreateRegistryInstance(ctx, createReq)

			errs[i] = createErr
		}(i, name)
	}

	wg.Wait()

	for i, name := range names {
		require.NoError(t, errs[i], "error creating registry instance %q", name)
	}

	for _, name := range names {
		waitForRegistryInstanceDeploy(t, env, name, true)
	}

	instanceA := findRegistryInstance(t, env, registryServiceName(containerRegistryCollisionInstanceNameA))
	instanceB := findRegistryInstance(t, env, registryServiceName(containerRegistryCollisionInstanceNameB))
	require.NotNil(t, instanceA)
	require.NotNil(t, instanceB)

	require.NotEqual(t, instanceA.GetPort(), instanceB.GetPort(), "registry instances share a container port")
	require.NotEqual(t, instanceA.GetUiPort(), instanceB.GetUiPort(), "registry instances share a ui port")

	inspectA, err := dockerClient.ContainerInspect(ctx, registryServiceName(containerRegistryCollisionInstanceNameA))
	require.NoError(t, err)
	require.True(t, inspectA.State.Running)

	inspectB, err := dockerClient.ContainerInspect(ctx, registryServiceName(containerRegistryCollisionInstanceNameB))
	require.NoError(t, err)
	require.True(t, inspectB.State.Running)

	inspectUiA, err := dockerClient.ContainerInspect(ctx, registryServiceName(containerRegistryCollisionInstanceNameA)+"-ui")
	require.NoError(t, err)
	require.True(t, inspectUiA.State.Running)

	inspectUiB, err := dockerClient.ContainerInspect(ctx, registryServiceName(containerRegistryCollisionInstanceNameB)+"-ui")
	require.NoError(t, err)
	require.True(t, inspectUiB.State.Running)
}

func Test_ContainerRegistryLifecycle(t *testing.T) {
	t.Parallel()
	RunPlaneSuite(t, Planes, func(plane Plane) suite.TestingSuite {
		return &ContainerRegistryLifecycleSuite{plane: plane}
	})
}

// newCreateRegistryInstanceRequest builds a CreateRegistryInstance.Request
// for the given instance name, opting into the UI sidecar only when
// enableUi is true (it defaults to off - see CreateRegistryInstance.Request.enable_ui).
func newCreateRegistryInstanceRequest(name string, enableUi bool) *velez_api.CreateRegistryInstance_Request {
	return &velez_api.CreateRegistryInstance_Request{Name: name, EnableUi: enableUi}
}

// findRegistryInstance returns the listed registry instance with exactly the
// given service name (registryServiceName for instances created by registryaas), or nil.
func findRegistryInstance(t *testing.T, env *TestEnvironment, name string) *velez_api.RegistryInstance {
	t.Helper()

	listReq := &velez_api.ListRegistryInstances_Request{}

	listResp, err := env.ContainerRegistryClient().ListRegistryInstances(t.Context(), listReq)
	require.NoError(t, err)

	for _, instance := range listResp.GetInstances() {
		if instance.GetName() == name {
			return instance
		}
	}

	return nil
}

// dropRegistryInstanceAndAwait calls DropRegistryInstance, which only enqueues
// drop_registry_instance, and waits for that task to finish.
func dropRegistryInstanceAndAwait(t *testing.T, env *TestEnvironment, serviceName string) {
	t.Helper()

	dropReq := &velez_api.DropRegistryInstance_Request{Name: serviceName}

	dropResp, err := env.ContainerRegistryClient().DropRegistryInstance(t.Context(), dropReq)
	require.NoError(t, err)
	require.Equal(t, serviceName, dropResp.GetEntityId())
	require.Equal(t, jobs.DropRegistryInstanceAction, dropResp.GetAction())

	dropTask := env.AwaitTask(t, dropResp.GetEntityId(), dropResp.GetAction())

	require.Equal(t, velez_api.TaskStatus_DONE, dropTask.GetStatus(),
		"drop_registry_instance task for %q error: %s", serviceName, dropTask.GetError())
}

// waitForRegistryInstanceDeploy first watches the create_registry_instance
// task itself (CreateRegistryInstance only enqueues it and returns - see
// ContainerRegistryLifecycleSuite's doc comment - so this is what proves the
// registry_instances/registries rows are actually written), then the
// create_smerd task(s) the deploy watcher dispatches for the registry
// instance and, when enableUi is true, its UI sidecar - proof either
// container actually exists, which the parent task alone doesn't give.
func waitForRegistryInstanceDeploy(t *testing.T, env *TestEnvironment, instanceName string, enableUi bool) {
	t.Helper()

	serviceName := registryServiceName(instanceName)

	instanceTask := env.AwaitTask(t, instanceName, jobs.CreateRegistryInstanceAction)

	require.Equal(t, velez_api.TaskStatus_DONE, instanceTask.GetStatus(),
		"create_registry_instance task for instance %q error: %s", instanceName, instanceTask.GetError())

	registryEntityID := jobs.SmerdEntityID("", serviceName)

	registryTask := env.AwaitTask(t, registryEntityID, jobs.CreateSmerdAction)

	require.Equal(t, velez_api.TaskStatus_DONE, registryTask.GetStatus(),
		"create_smerd task for registry instance %q error: %s", serviceName, registryTask.GetError())

	if !enableUi {
		return
	}

	uiEntityID := jobs.SmerdEntityID("", serviceName+"-ui")

	uiTask := env.AwaitTask(t, uiEntityID, jobs.CreateSmerdAction)

	require.Equal(t, velez_api.TaskStatus_DONE, uiTask.GetStatus(),
		"create_smerd task for registry ui sidecar %q error: %s", serviceName+"-ui", uiTask.GetError())
}

// removeContainerRegistryInstance force-removes the registry container, its UI
// sidecar container, and the per-instance data/auth volumes and network of the
// instance created under the caller-chosen name (all named behind
// labels.RegistryaasNamePrefix), ignoring "no such container/volume/network" -
// mirrors removePgaasInstance.
func removeContainerRegistryInstance(dockerClient client.APIClient, name string) {
	ctx := context.Background()
	removeOpts := container.RemoveOptions{Force: true}
	serviceName := registryServiceName(name)

	_ = dockerClient.ContainerRemove(ctx, serviceName, removeOpts)
	_ = dockerClient.ContainerRemove(ctx, serviceName+"-ui", removeOpts)
	_ = dockerClient.VolumeRemove(ctx, serviceName+"-data", true)
	_ = dockerClient.VolumeRemove(ctx, serviceName+"-auth", true)
	_ = dockerClient.NetworkRemove(ctx, serviceName+"-net")
}
