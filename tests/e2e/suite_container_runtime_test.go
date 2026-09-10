package e2e

import (
	"testing"

	"github.com/docker/docker/api/types/network"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

// Test_ContainerRuntime_Matrix covers the container-runtime abstraction
// (docs/container_runtimes) end to end: CreateSmerd through the real
// gRPC/transport path, then assertions against the live Docker daemon.
//
// It iterates the package-wide Plane matrix (matrix_test.go) - state backend
// x runtime backend - but only the single-node/docker cell is wired in
// Phase 1; see docs/container_runtimes/roadmap.md. Future cells (cluster/
// docker, a second runtime backend) join Planes rather than living as a
// suite-local matrix.
//
// Single-node (local_storage), no WithMatreshka(): Phase 1's only case needs
// neither cluster nor postgres, so it follows suite_environments_test.go's
// precedent rather than touching the shared matreshka fixture.
//
// containerRuntimeTestCase only carries what's specific to this suite's cases
// (which Plane, which environment) - the state/backend axes live on Plane
// itself.
type containerRuntimeTestCase struct {
	plane       Plane
	environment string
}

var containerRuntimeMatrix = []containerRuntimeTestCase{
	{plane: Planes[0], environment: "PROD"},
	// A cluster/docker case joins once Planes grows that cell - see
	// matrix_test.go.
}

const (
	// containerRuntimeSuffix is deliberately non-empty: the default
	// environment's suffix IS the node's ContainerSuffix, and with an empty one
	// "suffixed name" and "bare name" are the same string - the naming rule
	// under test would be unobservable.
	containerRuntimeSuffix = "e2ecrt"

	// Container names double as Docker hostnames (capped at 64 chars) and
	// GetServiceName(t) on a subtest of this matrix contains spaces and
	// slashes, so this case uses its own short, suite-unique name.
	containerRuntimeSmerdName = "e2e_ctr_runtime"

	networkIsoProdSuffix = "e2ecrtnet"
	networkIsoStageEnv   = "E2ECRTNETSTAGE"

	networkIsoProdName  = "e2e_ctr_net_prod"
	networkIsoStageName = "e2e_ctr_net_stage"

	listContainersProdSuffix = "e2ecrtlist"
	listContainersStageEnv   = "E2ECRTLISTSTAGE"

	listContainersProdName  = "e2e_ctr_runtime_list_prod"
	listContainersStageName = "e2e_ctr_runtime_list_stage"
)

// Left serial: every matrix row shares the fixed containerRuntimeSuffix /
// containerRuntimeSmerdName, so parallel rows (or a parallel parent racing a
// same-named container from another suite) would collide in Docker's global
// namespace. Cheap enough as one sequential case for now.
func Test_ContainerRuntime_Matrix(t *testing.T) {
	for _, tc := range containerRuntimeMatrix {
		t.Run(tc.plane.Name, func(t *testing.T) {
			runContainerRuntimeCase(t, tc)
		})
	}
}

// runContainerRuntimeCase stands up the fixture matching the case's Plane.
// Only single-node/docker is wired - Plane.NewEnvironment fails loudly for
// anything else rather than silently building the wrong fixture, so adding a
// matrix row without its fixture is impossible to miss.
func runContainerRuntimeCase(t *testing.T, tc containerRuntimeTestCase) {
	t.Helper()

	env := tc.plane.NewEnvironment(t, WithContainerSuffix(containerRuntimeSuffix))

	createReq := &velez_api.CreateSmerd_Request{
		Name:         containerRuntimeSmerdName,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
		Environment:  tc.environment,
	}

	// CreateSmerd is the task-watch path: the RPC enqueues the create_smerd
	// task and blocks until it reaches DONE/FAILED
	// (velez_api_impl/smerd_create.go), which is how every other e2e suite
	// waits for the jobs engine.
	created := env.CreateSmerd(t, createReq)
	require.Equal(t, velez_api.Smerd_running, created.GetStatus())

	// created.GetName() is the virtual/logical name - the ContainerRuntime
	// interface never surfaces the suffixed Docker name to callers (see
	// docs/container_runtimes/interface_design.md). expectedContainerName
	// computes the suffixed name separately, used below only to check
	// against the raw Docker daemon directly.
	require.Equal(t, containerRuntimeSmerdName, created.GetName())

	expectedName := expectedContainerName(containerRuntimeSmerdName, containerRuntimeSuffix)

	dockerClient := env.Custom.NodeClients.Docker().Client()

	inspected, err := dockerClient.ContainerInspect(t.Context(), expectedName)
	require.NoError(t, err, "expected a container named %q on the daemon", expectedName)

	require.Equal(t, "/"+expectedName, inspected.Name)
	require.Equal(t, created.GetUuid(), inspected.ID)
	require.Equal(t, containerRuntimeSuffix, inspected.Config.Labels[labels.SuffixLabel])
}

// expectedContainerName encodes the label-based runtime's name-conflict
// resolution: an empty suffix (the pre-multi-environment default) leaves the
// name untouched, a non-empty one appends "_<suffix>".
func expectedContainerName(name, suffix string) string {
	if suffix == "" {
		return name
	}

	return name + "_" + suffix
}

// Test_ContainerRuntime_ListContainers_ScopesToEnvironment is a RED test for
// the ContainerRuntime.ListContainers method that backs the ListSmerds RPC
// (see docs/container_runtimes/roadmap.md's Phase 1 and
// internal/clients/node_clients/container_runtime/label_based.go's
// ListContainers doc comment).
//
// It creates one smerd in the default (PROD) environment and a differently
// named one in a second environment, then asserts that a ListSmerds scoped to
// the second environment sees exactly its own container and none of PROD's.
//
// It is EXPECTED TO FAIL today: ContainerManager.ListSmerds now routes through
// RuntimeResolver.Runtime(...).ListContainers instead of calling
// node_clients.Docker.ListContainers directly, and labelBasedRuntime's
// ListContainers is a deliberate stub that does not filter by suffix (see its
// doc comment) - the next agent's job is to make it correct. Until then this
// scoped list returns BOTH environments' containers instead of just the one
// requested.
func Test_ContainerRuntime_ListContainers_ScopesToEnvironment(t *testing.T) {
	t.Parallel()

	env := NewEnvironment(t,
		WithContainerSuffix(listContainersProdSuffix),
		WithEnvironments([]string{listContainersStageEnv}))

	prodReq := &velez_api.CreateSmerd_Request{
		Name:         listContainersProdName,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
	}

	prodSmerd := env.CreateSmerd(t, prodReq)
	require.Equal(t, velez_api.Smerd_running, prodSmerd.GetStatus())

	stageReq := &velez_api.CreateSmerd_Request{
		Name:         listContainersStageName,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
		Environment:  listContainersStageEnv,
	}

	stageSmerd := env.CreateSmerd(t, stageReq)
	require.Equal(t, velez_api.Smerd_running, stageSmerd.GetStatus())

	stageListReq := &velez_api.ListSmerds_Request{
		Environment: listContainersStageEnv,
		Label:       map[string]string{testCaseNameLabel: t.Name()},
	}

	stageList := env.ListSmerds(t, t.Context(), stageListReq)

	require.Len(t, stageList.GetSmerds(), 1,
		"a STAGE-scoped ListSmerds must not also return PROD's container - "+
			"labelBasedRuntime.ListContainers is a deliberate unfiltered stub, see its doc comment")
	require.Equal(t, stageSmerd.GetUuid(), stageList.GetSmerds()[0].GetUuid())

	// Smerd.Name is always the virtual/logical name - never the suffixed
	// Docker container name - both from CreateSmerd's response (stageSmerd,
	// via InspectSmerd) and from ListSmerds (stageList, via
	// ContainerRuntime.ListContainers). See
	// docs/container_runtimes/interface_design.md.
	require.Equal(t, listContainersStageName, stageSmerd.GetName())
	require.Equal(t, listContainersStageName, stageList.GetSmerds()[0].GetName())
}

// Test_ContainerRuntime_Network_PerEnvironmentIsolation covers Stage 4
// (docs/container_runtimes): each environment must get its OWN Docker bridge
// network ("verv_<suffix>") instead of every environment sharing the single
// hardcoded "verv" network env.StartNetwork creates once at node boot. It
// creates one smerd in PROD and one differently-named smerd in a second
// environment - both with a port binding, so createContainerJob's
// default-network connect actually runs (see label_based.go's
// CreateNetwork/ConnectToNetwork) - then asserts against the real Docker
// daemon that each container ends up on its own suffixed network and that
// neither network contains the other environment's container.
//
// This used to be RED for a structural reason beyond "the methods don't
// exist yet": dockerutils.CreateNetwork hardcoded the exact same IPAM subnet
// (10.0.1.0/24) for every network it created, regardless of name, which was
// fine when only one such network (the single shared "verv") ever existed on
// a node but made a second environment's own "verv_<suffix>" network collide
// with it ("Pool overlaps with other one on this address space"). Fixed by
// dropping the explicit Subnet from dockerutils.CreateNetwork's IPAM config
// and letting Docker's default IPAM driver auto-allocate a non-overlapping
// subnet per network - env.StartNetwork's existing "verv" network is
// untouched by that change (CreateNetwork no-ops when a network with the
// requested name already exists).
func Test_ContainerRuntime_Network_PerEnvironmentIsolation(t *testing.T) {
	t.Parallel()

	env := NewEnvironment(t,
		WithContainerSuffix(networkIsoProdSuffix),
		WithEnvironments([]string{networkIsoStageEnv}))

	prodReq := &velez_api.CreateSmerd_Request{
		Name:         networkIsoProdName,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
		Settings: &velez_api.Container_Settings{
			Ports: []*velez_api.Port{{ServicePortNumber: 8080, Protocol: velez_api.Port_tcp}},
		},
	}

	prodSmerd := env.CreateSmerd(t, prodReq)
	require.Equal(t, velez_api.Smerd_running, prodSmerd.GetStatus())
	checkPorts(t, prodSmerd.GetPorts(), prodReq.GetSettings().GetPorts())

	stageReq := &velez_api.CreateSmerd_Request{
		Name:         networkIsoStageName,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
		Environment:  networkIsoStageEnv,
		Settings: &velez_api.Container_Settings{
			Ports: []*velez_api.Port{{ServicePortNumber: 8080, Protocol: velez_api.Port_tcp}},
		},
	}

	stageSmerd := env.CreateSmerd(t, stageReq)
	require.Equal(t, velez_api.Smerd_running, stageSmerd.GetStatus())
	checkPorts(t, stageSmerd.GetPorts(), stageReq.GetSettings().GetPorts())

	dockerClient := env.Custom.NodeClients.Docker().Client()

	prodNetName := "verv_" + networkIsoProdSuffix
	stageNetName := "verv_" + networkIsoStageEnv

	inspectOpts := network.InspectOptions{Verbose: true}

	prodNet, err := dockerClient.NetworkInspect(t.Context(), prodNetName, inspectOpts)
	require.NoError(t, err, "expected PROD's own network %q to exist", prodNetName)

	stageNet, err := dockerClient.NetworkInspect(t.Context(), stageNetName, inspectOpts)
	require.NoError(t, err, "expected STAGE's own network %q to exist", stageNetName)

	_, prodInProdNet := prodNet.Containers[prodSmerd.GetUuid()]
	require.True(t, prodInProdNet, "expected PROD's container to be connected to %q", prodNetName)

	_, stageInProdNet := prodNet.Containers[stageSmerd.GetUuid()]
	require.False(t, stageInProdNet, "STAGE's container must not be connected to PROD's network")

	_, stageInStageNet := stageNet.Containers[stageSmerd.GetUuid()]
	require.True(t, stageInStageNet, "expected STAGE's container to be connected to %q", stageNetName)

	_, prodInStageNet := stageNet.Containers[prodSmerd.GetUuid()]
	require.False(t, prodInStageNet, "PROD's container must not be connected to STAGE's network")
}
