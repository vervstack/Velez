package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	// containerRuntimeSuffix is deliberately non-empty: the default
	// environment's suffix IS the node's ContainerSuffix, and with an empty
	// one "suffixed name" and "bare name" are the same string - the naming
	// rule under test would be unobservable.
	containerRuntimeSuffix = "e2ecrt"

	// containerRuntimeSmerdName is the shared base name every environment
	// case creates its variant(s) under - shared on purpose, see
	// Test_ContainerRuntime_Matrix's doc comment.
	containerRuntimeSmerdName = "e2e_ctr_runtime"

	containerRuntimeStageEnv = "E2ECRTSTAGE"
)

// environmentCase is the outer table: one environment to exercise within a
// Plane's fixture.
//
//   - environment is the velez_api.CreateSmerd_Request.Environment value -
//     "" means the node's own/default environment.
//   - suffix is what labelSuffixResolver actually stamps real Docker names
//     with for it: the default environment's suffix is whatever
//     ContainerSuffix the fixture was built with (containerRuntimeSuffix),
//     every other registered environment's suffix is its own name (see
//     internal/storage/environments/static.go's NewStatic).
type environmentCase struct {
	name        string
	environment string
	suffix      string
}

// containerVariant is the inner table: one container shape to create and
// verify per environment. check runs after the common assertions
// (runContainerRuntimeCase) already passed - nil means no extra assertions
// beyond those, giving a future variant (wget nginx, `SELECT 1` over the
// postgres driver, ...) a new table row instead of a new test function.
type containerVariant struct {
	// name suffixes the smerd's logical name, so multiple variants in the
	// same environment don't collide in Docker's global container namespace.
	name      string
	imageName string
	settings  *velez_api.Container_Settings
	check     func(t *testing.T, env *TestEnvironment, created *velez_api.Smerd)
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
// the live Docker daemon that a container actually exists. It stays untagged
// (no //go:build) so it keeps running in every `make test-e2e` invocation,
// local and CI - see suite_container_runtime_scoping_test.go for more
// involved cases, gated behind the e2e_full build tag.
//
// It is a three-axis matrix - Plane x environment x container variant - built
// as one function per axis (Test_ContainerRuntime_Matrix ->
// runContainerRuntimePlane -> runContainerRuntimeEnvironmentCase ->
// runContainerRuntimeCase) so no single function juggles more than one loop,
// while the t.Run tree still lets you target one cell directly
// (-run Matrix/single-node.docker/PROD/hello-world).
//
// Every environmentCase shares one smerdName on purpose: that is what proves
// labelSuffixResolver's guarantee that Smerd.Name stays env-agnostic even
// though the real Docker container name differs per environment (see
// runContainerRuntimeCase).
func Test_ContainerRuntime_Matrix(t *testing.T) {
	t.Parallel()

	for _, plane := range containerRuntimePlanes {
		t.Run(plane.Name, func(t *testing.T) {
			runContainerRuntimePlane(t, plane)
		})
	}
}

// runContainerRuntimePlane builds the ONE fixture for plane - registering
// every non-default environment the table needs, derived from
// containerRuntimeEnvironments itself so a new row can't drift out of sync
// with what the fixture actually registers - then iterates the environment
// axis against it.
func runContainerRuntimePlane(t *testing.T, plane Plane) {
	t.Helper()
	t.Parallel()

	env := plane.NewEnvironment(t,
		WithContainerSuffix(containerRuntimeSuffix),
		WithEnvironments(containerRuntimeExtraEnvironments()))

	for _, envCase := range containerRuntimeEnvironments {
		t.Run(envCase.name, func(t *testing.T) {
			runContainerRuntimeEnvironmentCase(t, env, envCase)
		})
	}
}

// containerRuntimeExtraEnvironments collects every non-default environment
// name out of containerRuntimeEnvironments, in the shape WithEnvironments
// wants - so adding a table row is the only edit needed to also register it
// on the fixture.
func containerRuntimeExtraEnvironments() []string {
	var envs []string

	for _, envCase := range containerRuntimeEnvironments {
		if envCase.environment != "" {
			envs = append(envs, envCase.environment)
		}
	}

	return envs
}

// runContainerRuntimeEnvironmentCase iterates the container-variant axis for
// one environment.
func runContainerRuntimeEnvironmentCase(t *testing.T, env *TestEnvironment, envCase environmentCase) {
	t.Helper()
	t.Parallel()

	for _, variant := range containerRuntimeVariants {
		t.Run(variant.name, func(t *testing.T) {
			runContainerRuntimeCase(t, env, envCase, variant)
		})
	}
}

// runContainerRuntimeCase creates variant in envCase's environment and
// checks it: the assertions every variant needs (requireSmerdRunning,
// requireLogicalName, requireRealContainer), then variant's own optional
// check.
func runContainerRuntimeCase(t *testing.T, env *TestEnvironment, envCase environmentCase, variant containerVariant) {
	t.Helper()
	t.Parallel()

	smerdName := containerRuntimeSmerdName + "_" + variant.name
	createReq := newContainerRuntimeCreateRequest(smerdName, envCase.environment, variant.imageName, variant.settings)

	// CreateSmerd is the task-watch path: the RPC enqueues the create_smerd
	// task and blocks until it reaches DONE/FAILED
	// (velez_api_impl/smerd_create.go), which is how every other e2e suite
	// waits for the jobs engine.
	created := env.CreateSmerd(t, createReq)

	requireSmerdRunning(t, created)
	requireLogicalName(t, created, smerdName)
	requireRealContainer(t, env, created, smerdName, envCase.suffix)

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

func requireSmerdRunning(t *testing.T, created *velez_api.Smerd) {
	t.Helper()

	require.Equal(t, velez_api.Smerd_running, created.GetStatus())
}

// requireLogicalName is the actual proof of labelSuffixResolver's guarantee:
// created.GetName() must be the SAME env-agnostic name regardless of which
// environment created it - the ContainerRuntime interface never surfaces the
// suffixed Docker name to callers (see
// docs/container_runtimes/interface_design.md).
func requireLogicalName(t *testing.T, created *velez_api.Smerd, name string) {
	t.Helper()

	require.Equal(t, name, created.GetName())
}

// requireRealContainer asserts against the live Docker daemon that the real
// container name/labels carry the suffix logicalName's environment maps to,
// even though created.GetName() (checked separately by requireLogicalName)
// does not.
func requireRealContainer(t *testing.T, env *TestEnvironment, created *velez_api.Smerd, logicalName, suffix string) {
	t.Helper()

	expectedName := expectedContainerName(logicalName, suffix)

	dockerClient := env.Custom.NodeClients.Docker().Client()

	inspected, err := dockerClient.ContainerInspect(t.Context(), expectedName)
	require.NoError(t, err, "expected a container named %q on the daemon", expectedName)

	require.Equal(t, "/"+expectedName, inspected.Name)
	require.Equal(t, created.GetUuid(), inspected.ID)
	require.Equal(t, suffix, inspected.Config.Labels[labels.SuffixLabel])
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
