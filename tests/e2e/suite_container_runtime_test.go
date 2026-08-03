package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

// Test_ContainerRuntime_Matrix covers the container-runtime abstraction
// (docs/container_runtimes) end to end: CreateSmerd through the real
// gRPC/transport path, then assertions against the live Docker daemon.
//
// It is a 2x2 matrix - state backend (none/postgres) x runtime backend
// (label-based/dedicated) - but only the no-state/label-based cell is wired in
// Phase 1; see docs/container_runtimes/roadmap.md. The other three cells are
// listed as commented-out placeholders instead of skipped subtests so the
// matrix documents what is next without pretending to cover it.
//
// Single-node (local_storage), no WithMatreshka(): Phase 1's only case needs
// neither cluster nor postgres, so it follows suite_environments_test.go's
// precedent rather than touching the shared matreshka fixture.
type containerRuntimeTestCase struct {
	name           string
	stateMode      string // "none" | "postgres"
	runtimeBackend string // "label" | "dedicated"
	environment    string
}

var containerRuntimeMatrix = []containerRuntimeTestCase{
	{name: "no-state/label-based (default)", stateMode: "none", runtimeBackend: "label", environment: "PROD"},
	// {stateMode: "postgres", runtimeBackend: "label"}      — next
	// {stateMode: "none",     runtimeBackend: "dedicated"}  — next
	// {stateMode: "postgres", runtimeBackend: "dedicated"}  — next
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

	containerRuntimeStateNone     = "none"
	containerRuntimeBackendLabels = "label"

	listContainersProdSuffix = "e2ecrtlist"
	listContainersStageEnv   = "E2ECRTLISTSTAGE"

	listContainersProdName  = "e2e_ctr_runtime_list_prod"
	listContainersStageName = "e2e_ctr_runtime_list_stage"
)

func Test_ContainerRuntime_Matrix(t *testing.T) {
	for _, tc := range containerRuntimeMatrix {
		t.Run(tc.name, func(t *testing.T) {
			runContainerRuntimeCase(t, tc)
		})
	}
}

// runContainerRuntimeCase stands up the fixture matching the case's
// stateMode/runtimeBackend. Only none/label is wired - the others fail loudly
// rather than silently skipping, so adding a matrix row without its fixture is
// impossible to miss.
func runContainerRuntimeCase(t *testing.T, tc containerRuntimeTestCase) {
	t.Helper()

	if tc.stateMode != containerRuntimeStateNone {
		t.Fatalf("add fixture support for stateMode %q", tc.stateMode)
	}

	if tc.runtimeBackend != containerRuntimeBackendLabels {
		t.Fatalf("add fixture support for runtimeBackend %q", tc.runtimeBackend)
	}

	env := NewEnvironment(t, WithContainerSuffix(containerRuntimeSuffix))

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

	expectedName := expectedContainerName(containerRuntimeSmerdName, containerRuntimeSuffix)
	require.Equal(t, expectedName, created.GetName())

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

	// Open design question (observe, don't fix - see the task this test was
	// written for): Smerd.Name currently surfaces the raw, suffixed Docker
	// container name rather than the logical name CreateSmerd was called
	// with. container_manager/smerd_list.go does
	// `smerd.Name = container.Names[0][1:]` verbatim.
	require.Equal(t, expectedContainerName(listContainersStageName, listContainersStageEnv),
		stageSmerd.GetName(),
		"observation: Smerd.Name is the suffixed Docker name, not the bare logical name")
}
