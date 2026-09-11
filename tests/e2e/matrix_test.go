package e2e

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

// PlaneMode is the state backend a Plane's fixture runs with.
type PlaneMode string

// PlaneBackend is the container-runtime backend a Plane exercises. Only the
// label-based ContainerRuntime backend (a shared Docker daemon) is
// implemented today - see docs/container_runtimes/roadmap.md.
type PlaneBackend string

// EnvSeparationWay is how a Plane's environments (PROD/STAGE/...) are kept
// from colliding on the same Docker engine.
type EnvSeparationWay string

// RunningMode is how the Velez app under test is actually run.
type RunningMode string

const (
	ModeSingleNode PlaneMode = "single-node"
	ModeCluster    PlaneMode = "cluster"

	BackendDocker PlaneBackend = "docker"

	// SeparationLabelBased: one shared Docker engine, environments kept
	// apart by name/label/suffix (labelSuffixResolver). The only wired value
	// today.
	SeparationLabelBased EnvSeparationWay = "label-based"

	// SeparationDedicatedEngine: each environment gets its own Docker
	// daemon/host (domain.Environment.DockerHost). NOT implemented -
	// RuntimeResolver.Runtime returns ErrDedicatedRuntimeNotImplemented for
	// any environment with a non-empty DockerHost today. See
	// https://trello.com/c/otriSswo.
	SeparationDedicatedEngine EnvSeparationWay = "separate-engine"

	// RunningModeBinary: the in-process app fixture NewEnvironment builds
	// today (bufconn, no real container). The only wired value today.
	RunningModeBinary RunningMode = "binary"

	// RunningModeContainer: a real Velez container, docker-socket-mounted,
	// configured via env vars, driven over a real network connection instead
	// of bufconn. NOT implemented - see https://trello.com/c/otriSswo.
	RunningModeContainer RunningMode = "container"
)

// Plane names one cell of the package-wide fixture-shape matrix, shared
// across every suite in this package. A suite adopts it by iterating Planes
// and building its fixture through Plane.NewEnvironment instead of calling
// NewEnvironment directly, so a new cell (a cluster fixture, a dedicated
// engine, a containerized Velez) lights up in every adopting suite at once
// instead of needing a per-suite matrix rewrite.
//
// There is no separate display Name: Plane.Name() derives it from the axis
// values themselves, so two cells can never silently share a label and a
// t.Run tree/log line always says exactly what ran.
type Plane struct {
	Mode       PlaneMode
	Backend    PlaneBackend
	Separation EnvSeparationWay
	Running    RunningMode
}

// Name is this Plane's t.Run/log identity, generated from its axis values.
func (p Plane) Name() string {
	return strings.Join([]string{string(p.Mode), string(p.Backend), string(p.Separation), string(p.Running)}, "/")
}

// Planes is the package-wide matrix of fixture cells. Only
// single-node/docker/label-based/binary is wired to a real fixture today -
// see Plane.NewEnvironment.
var Planes = []Plane{
	{
		Mode:       ModeSingleNode,
		Backend:    BackendDocker,
		Separation: SeparationLabelBased,
		Running:    RunningModeBinary,
	},

	// cluster/docker: needs WithMatreshka()+WithClusterPgDsn threaded through
	// per-suite before it can be a live cell here - see
	// docs/container_runtimes/roadmap.md. Not wired yet, deliberately.

	// single-node/docker cells with Separation: SeparationDedicatedEngine or
	// Running: RunningModeContainer: both need real product/infra work first
	// - see https://trello.com/c/otriSswo. Not wired yet, deliberately.
}

// NewEnvironment builds the TestEnvironment fixture matching p's Mode,
// Backend, Separation and Running. Today all four are single-case, so this
// is a passthrough to package-level NewEnvironment. An unimplemented
// combination skips the calling test rather than silently building the
// wrong fixture, so adding a matrix row without wiring its fixture here is
// impossible to miss.
func (p Plane) NewEnvironment(t *testing.T, opts ...TestEnvOpt) *TestEnvironment {
	t.Helper()

	if p.Mode != ModeSingleNode {
		t.Skipf("plane %q: mode %q not wired to a fixture yet", p.Name(), p.Mode)
	}

	if p.Backend != BackendDocker {
		t.Skipf("plane %q: backend %q not wired to a fixture yet", p.Name(), p.Backend)
	}

	if p.Separation != SeparationLabelBased {
		t.Skipf("plane %q: env-separation-way %q not wired to a fixture yet - see https://trello.com/c/otriSswo",
			p.Name(), p.Separation)
	}

	if p.Running != RunningModeBinary {
		t.Skipf("plane %q: running-mode %q not wired to a fixture yet - see https://trello.com/c/otriSswo",
			p.Name(), p.Running)
	}

	return NewEnvironment(t, opts...)
}

// RunPlaneSuite runs newSuite once per plane in planes, each as its own
// t.Run(plane.Name(), ...) subtest. It is how a testify suite adopts the
// Plane matrix: build the suite's fixture through s.plane.NewEnvironment
// (inside SetupTest/SetupSuite or a test method) instead of the suite
// iterating Planes itself. Plane.NewEnvironment's t.Skipf guards remain the
// only skip mechanism - RunPlaneSuite adds none of its own, so an
// unimplemented cell still reports as skipped rather than silently absent.
func RunPlaneSuite(t *testing.T, planes []Plane, newSuite func(Plane) suite.TestingSuite) {
	t.Helper()

	for _, plane := range planes {
		t.Run(plane.Name(), func(t *testing.T) {
			t.Parallel()
			suite.Run(t, newSuite(plane))
		})
	}
}
