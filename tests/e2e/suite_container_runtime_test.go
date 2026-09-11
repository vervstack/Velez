package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	// containerRuntimeSuffix is deliberately non-empty: the default
	// environment's suffix IS the node's ContainerSuffix, and with an empty one
	// "suffixed name" and "bare name" are the same string - the naming rule
	// under test would be unobservable.
	containerRuntimeSuffix = "e2ecrt"

	// containerRuntimeSmerdName is the shared base name every environment
	// case creates its variant(s) under - shared on purpose, see
	// Test_ContainerRuntime_Matrix's doc comment. Container names double as
	// Docker hostnames (capped at 64 chars) and GetServiceName(t) on a
	// subtest of this matrix contains spaces and slashes, so this suite uses
	// its own short, unique name instead.
	containerRuntimeSmerdName = "e2e_ctr_runtime"

	containerRuntimeStageEnv = "E2ECRTSTAGE"
)

// environmentCase is the outer table: one environment to exercise within a
// Plane's fixture.
//
//   - environment is the velez_api.CreateSmerd_Request.Environment value -
//     "" means the node's own/default environment.
//   - suffix is what labelSuffixResolver actually stamps real Docker names
//     with for it. Mirrors internal/storage/environments/static.go's
//     NewStatic: the default environment's suffix is whatever
//     ContainerSuffix the fixture was built with (containerRuntimeSuffix
//     above), every other registered environment's suffix is its own name.
//     Kept here explicitly rather than re-derived, so a table row states
//     its expectation instead of relying on a reader already knowing the
//     production naming rule.
type environmentCase struct {
	name        string
	environment string
	suffix      string
}

// containerVariant is the inner table: one container shape to create and
// verify per environment. Today there is a single "hello world, no extra
// checks" variant - check exists so a future variant (wget nginx, `SELECT 1`
// over the postgres driver, an exec'd healthcheck, etc.) is a new table row
// instead of a new test function.
type containerVariant struct {
	// name suffixes the smerd's logical name, so multiple variants in the
	// same environment don't collide in Docker's global container namespace.
	name      string
	imageName string
	settings  *velez_api.Container_Settings
	// check runs after the smerd's Smerd_running status and the real-Docker
	// name/label assertions common to every variant (runContainerRuntimeCase)
	// already passed. nil means no extra assertions beyond those.
	check func(t *testing.T, env *TestEnvironment, created *velez_api.Smerd)
}

var (
	// containerRuntimePlanes: only single-node/docker (Planes[0]) is wired
	// for now; a cluster/docker cell joins here once Plane.NewEnvironment
	// supports it - see matrix_test.go.
	containerRuntimePlanes = []Plane{Planes[0]}

	containerRuntimeEnvironments = []environmentCase{
		{name: "PROD", environment: "", suffix: containerRuntimeSuffix},
		{name: "STAGE", environment: containerRuntimeStageEnv, suffix: containerRuntimeStageEnv},
	}

	containerRuntimeVariants = []containerVariant{
		{name: "hello-world", imageName: HelloWorldAppImage},
	}
)

// Test_ContainerRuntime_Matrix is the package's always-on smoke-deploy test:
// CreateSmerd through the real gRPC/transport path, then assertions against
// the live Docker daemon that a container actually exists. It stays
// untagged (no //go:build) so it keeps running in every `make test-e2e`
// invocation, local and CI - see suite_container_runtime_scoping_test.go for
// more involved cases, gated behind the e2e_full build tag.
//
// It is a two-level table: the outer level (containerRuntimeEnvironments)
// drives which environment a case runs in, the inner level
// (containerRuntimeVariants) drives which container gets created and
// checked. Both levels run inside ONE fixture per Plane (one node, several
// registered environments - mirrors production: a node has one Docker
// daemon shared by every environment on it), and share the same logical
// smerd name across environments on purpose: that is what actually proves
// labelSuffixResolver's guarantee that Smerd.Name is env-agnostic even
// though the real Docker container name differs per environment (see
// runContainerRuntimeCase).
//
// Left serial for now: environments/variants land on distinct real Docker
// names (different suffix or different variant name), so nothing here
// actually collides - but this is a shared, always-on smoke test, and
// t.Parallel() buys little at this table's current size. Revisit once the
// table grows enough for it to matter.
func Test_ContainerRuntime_Matrix(t *testing.T) {
	for _, plane := range containerRuntimePlanes {
		t.Run(plane.Name, func(t *testing.T) {
			runContainerRuntimePlane(t, plane)
		})
	}
}

// runContainerRuntimePlane builds the ONE fixture for plane - registering
// every environment the outer table needs up front - then iterates the outer
// (environment) and inner (variant) tables against it.
func runContainerRuntimePlane(t *testing.T, plane Plane) {
	t.Helper()

	env := plane.NewEnvironment(t,
		WithContainerSuffix(containerRuntimeSuffix),
		WithEnvironments([]string{containerRuntimeStageEnv}))

	for _, envCase := range containerRuntimeEnvironments {
		t.Run(envCase.name, func(t *testing.T) {
			for _, variant := range containerRuntimeVariants {
				t.Run(variant.name, func(t *testing.T) {
					runContainerRuntimeCase(t, env, envCase, variant)
				})
			}
		})
	}
}

// runContainerRuntimeCase creates variant in envCase's environment and
// checks it two ways: assertions every variant needs (Smerd_running, the
// env-agnostic logical Name, the real suffixed Docker name/labels), then
// variant's own optional check.
func runContainerRuntimeCase(t *testing.T, env *TestEnvironment, envCase environmentCase, variant containerVariant) {
	t.Helper()

	smerdName := containerRuntimeSmerdName + "_" + variant.name

	createReq := newContainerRuntimeCreateRequest(smerdName, envCase.environment, variant.imageName, variant.settings)

	// CreateSmerd is the task-watch path: the RPC enqueues the create_smerd
	// task and blocks until it reaches DONE/FAILED
	// (velez_api_impl/smerd_create.go), which is how every other e2e suite
	// waits for the jobs engine.
	created := env.CreateSmerd(t, createReq)
	require.Equal(t, velez_api.Smerd_running, created.GetStatus())

	// created.GetName() must be the SAME env-agnostic logical name
	// regardless of envCase - the ContainerRuntime interface never surfaces
	// the suffixed Docker name to callers (see
	// docs/container_runtimes/interface_design.md). This is the actual proof
	// of that guarantee: every environmentCase in the outer table shares one
	// smerdName, so PROD and STAGE only diverge in the real Docker name
	// checked below, never in what CreateSmerd/InspectSmerd report back.
	require.Equal(t, smerdName, created.GetName())

	expectedName := expectedContainerName(smerdName, envCase.suffix)

	dockerClient := env.Custom.NodeClients.Docker().Client()

	inspected, err := dockerClient.ContainerInspect(t.Context(), expectedName)
	require.NoError(t, err, "expected a container named %q on the daemon", expectedName)

	require.Equal(t, "/"+expectedName, inspected.Name)
	require.Equal(t, created.GetUuid(), inspected.ID)
	require.Equal(t, envCase.suffix, inspected.Config.Labels[labels.SuffixLabel])

	if variant.check != nil {
		variant.check(t, env, created)
	}
}

// newContainerRuntimeCreateRequest builds a CreateSmerd_Request for one
// (environment, variant) pair - named out so every case in the matrix
// assigns fields to a named variable instead of constructing the request
// literal inline at the call site.
func newContainerRuntimeCreateRequest(
	name, environment, imageName string, settings *velez_api.Container_Settings,
) *velez_api.CreateSmerd_Request {
	req := &velez_api.CreateSmerd_Request{
		Name:         name,
		ImageName:    imageName,
		IgnoreConfig: true,
		Environment:  environment,
		Settings:     settings,
	}

	return req
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
