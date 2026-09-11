package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

// Test_ContainerRuntime_Matrix is the package's always-on smoke-deploy test:
// CreateSmerd through the real gRPC/transport path, then assertions against
// the live Docker daemon that a container actually exists. It stays
// untagged (no //go:build) so it keeps running in every `make test-e2e`
// invocation, local and CI - see suite_container_runtime_scoping_test.go for
// the suite's more involved multi-environment cases, which are gated behind
// the e2e_full build tag instead.
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
