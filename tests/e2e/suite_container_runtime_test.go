package e2e

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/storage/environments"
)

// StateMode is the Velez state backend a case runs the app against.
type StateMode string

const (
	// containerRuntimeSmerdName is the shared base name every environment
	// case creates its variant(s) under - shared on purpose, see
	// Test_ContainerRuntime_Matrix's doc comment.
	containerRuntimeSmerdName = "e2e_ctr_runtime"

	containerRuntimeStageEnvName = "STAGE"

	StateModeStateless StateMode = "stateless"

	// StateModeStatefull: a real postgres plugin instance
	// (enableStatefullPgUnderDind, tests/e2e/helper_statefull_test.go). NOT
	// wired into this always-on, fully-parallel smoke matrix yet -
	// enableStatefullPgUnderDind must t.Chdir to the repo root (goose
	// migrations resolve "./migrations" relatively) and explicitly forbids
	// t.Parallel(), both incompatible with this suite's shared, parallel
	// per-Plane fixture today. See https://trello.com/c/otriSswo.
	StateModeStatefull StateMode = "statefull"
)

// containerVariant is one container shape to create and verify. check runs
// after the common assertions (runContainerRuntimeCase) already passed - nil
// means no extra assertions beyond those, giving a future variant (wget
// nginx, `SELECT 1` over the postgres driver, ...) a new table row instead of
// a new test function.
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

	// containerRuntimeEnvironments is both the velez_api.CreateSmerd_Request
	// Environment value AND this environment's t.Run/display name - the two
	// were never independent data (an environment's name IS what identifies
	// it on the wire), so there is only one field. environments.
	// DefaultEnvironmentName ("PROD") is the real production sentinel for
	// the node's own/default environment - internal/storage/environments/
	// static.go always seeds a row under exactly that name, so sending it
	// explicitly resolves identically to sending "" (see environments.
	// Resolve's doc comment); using it here instead of "" is what lets this
	// table avoid a second field.
	containerRuntimeEnvironments = []string{
		environments.DefaultEnvironmentName,
		containerRuntimeStageEnvName,
	}

	// containerRuntimeStateModes: only StateModeStateless is wired today -
	// see its doc comment.
	containerRuntimeStateModes = []StateMode{StateModeStateless}

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
// It is a matrix over Plane (matrix_test.go: single-node/cluster x
// docker/... backend x env-separation-way x running-mode) x environment x
// state-mode x container variant, built as one function per axis
// (Test_ContainerRuntime_Matrix -> runContainerRuntimePlane ->
// runContainerRuntimeEnvironmentCase -> runContainerRuntimeStateModeCase ->
// runContainerRuntimeCase) so no single function juggles more than one loop,
// while the t.Run tree still lets you target one cell directly
// (-run Matrix/single-node.docker.../PROD/stateless/hello-world).
//
// Every environment case shares one smerdName on purpose: that is what
// proves labelSuffixResolver's guarantee that Smerd.Name stays env-agnostic
// even though the real Docker container name differs per environment (see
// runContainerRuntimeCase).
func Test_ContainerRuntime_Matrix(t *testing.T) {
	t.Parallel()

	for _, plane := range containerRuntimePlanes {
		t.Run(plane.Name(), func(t *testing.T) {
			runContainerRuntimePlane(t, plane)
		})
	}
}

// runContainerRuntimePlane builds the ONE fixture for plane. nodeSuffix is
// this fixture's ContainerSuffix (the default/PROD environment's suffix,
// see environmentSuffix) - derived from t.Name() rather than a hand-picked
// constant, so it is unique per test run by construction instead of by every
// suite in this package remembering to pick a distinct string. Every
// non-default environment the environment axis needs is registered up
// front, derived from containerRuntimeEnvironments itself so a new row can't
// drift out of sync with what the fixture actually registers.
func runContainerRuntimePlane(t *testing.T, plane Plane) {
	t.Helper()
	t.Parallel()

	nodeSuffix := dockerSafeToken(t.Name())

	env := plane.NewEnvironment(t,
		WithContainerSuffix(nodeSuffix),
		WithEnvironments(containerRuntimeExtraEnvironments()))

	for _, environment := range containerRuntimeEnvironments {
		t.Run(environment, func(t *testing.T) {
			runContainerRuntimeEnvironmentCase(t, env, environment, nodeSuffix)
		})
	}
}

// dockerSafeToken turns s (a t.Name(), which contains "/" between subtest
// levels) into a string safe to append to a Docker container name.
func dockerSafeToken(s string) string {
	return strings.NewReplacer("/", "_", " ", "_").Replace(s)
}

// containerRuntimeExtraEnvironments collects every non-default environment
// name out of containerRuntimeEnvironments, in the shape WithEnvironments
// wants - so adding a table row is the only edit needed to also register it
// on the fixture. environments.DefaultEnvironmentName is excluded: it's
// already seeded by NewStatic regardless, and re-listing it is a no-op there
// (see static.go's NewStatic).
func containerRuntimeExtraEnvironments() []string {
	var envs []string

	for _, environment := range containerRuntimeEnvironments {
		if environment != environments.DefaultEnvironmentName {
			envs = append(envs, environment)
		}
	}

	return envs
}

// runContainerRuntimeEnvironmentCase iterates the state-mode axis for one
// environment.
func runContainerRuntimeEnvironmentCase(t *testing.T, env *TestEnvironment, environment, nodeSuffix string) {
	t.Helper()
	t.Parallel()

	for _, stateMode := range containerRuntimeStateModes {
		t.Run(string(stateMode), func(t *testing.T) {
			runContainerRuntimeStateModeCase(t, env, environment, nodeSuffix)
		})
	}
}

// runContainerRuntimeStateModeCase iterates the container-variant axis.
// stateMode itself isn't threaded any further today - the only wired value
// (StateModeStateless) needs no setup beyond what NewEnvironment already
// does - it exists purely as a matrix/t.Run axis until StateModeStatefull
// is wired (see its doc comment).
func runContainerRuntimeStateModeCase(t *testing.T, env *TestEnvironment, environment, nodeSuffix string) {
	t.Helper()
	t.Parallel()

	for _, variant := range containerRuntimeVariants {
		t.Run(variant.name, func(t *testing.T) {
			runContainerRuntimeCase(t, env, environment, nodeSuffix, variant)
		})
	}
}

// runContainerRuntimeCase creates variant in environment and checks it: the
// assertions every variant needs (requireSmerdRunning, requireLogicalName,
// requireRealContainer), then variant's own optional check.
func runContainerRuntimeCase(
	t *testing.T, env *TestEnvironment, environment, nodeSuffix string, variant containerVariant,
) {
	t.Helper()
	t.Parallel()

	smerdName := containerRuntimeSmerdName + "_" + variant.name
	createReq := newContainerRuntimeCreateRequest(smerdName, environment, variant.imageName, variant.settings)

	// CreateSmerd is the task-watch path: the RPC enqueues the create_smerd
	// task and blocks until it reaches DONE/FAILED
	// (velez_api_impl/smerd_create.go), which is how every other e2e suite
	// waits for the jobs engine.
	created := env.CreateSmerd(t, createReq)

	requireSmerdRunning(t, created)
	requireLogicalName(t, created, smerdName)
	requireRealContainer(t, env, created, smerdName, environmentSuffix(nodeSuffix, environment))

	if variant.check != nil {
		variant.check(t, env, created)
	}
}

// environmentSuffix mirrors internal/storage/environments/static.go's
// NewStatic: the default (PROD) environment's suffix is whatever
// ContainerSuffix the fixture was built with (nodeSuffix), every other
// registered environment's suffix is its own name.
func environmentSuffix(nodeSuffix, environment string) string {
	if environment == environments.DefaultEnvironmentName {
		return nodeSuffix
	}

	return environment
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
