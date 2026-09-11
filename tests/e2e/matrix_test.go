package e2e

import (
	"testing"
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
type Plane struct {
	Name       string
	Mode       PlaneMode
	Backend    PlaneBackend
	Separation EnvSeparationWay
	Running    RunningMode
}

// Planes is the package-wide matrix of fixture cells. Only
// single-node/docker/label-based/binary is wired to a real fixture today -
// see Plane.NewEnvironment.
var Planes = []Plane{
	{
		Name:       "single-node/docker",
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
// combination fails loudly rather than silently building the wrong fixture,
// so adding a matrix row without wiring its fixture here is impossible to
// miss.
func (p Plane) NewEnvironment(t *testing.T, opts ...TestEnvOpt) *TestEnvironment {
	t.Helper()

	if p.Mode != ModeSingleNode {
		t.Fatalf("plane %q: mode %q not wired to a fixture yet", p.Name, p.Mode)
	}

	if p.Backend != BackendDocker {
		t.Fatalf("plane %q: backend %q not wired to a fixture yet", p.Name, p.Backend)
	}

	if p.Separation != SeparationLabelBased {
		t.Fatalf("plane %q: env-separation-way %q not wired to a fixture yet - see https://trello.com/c/otriSswo",
			p.Name, p.Separation)
	}

	if p.Running != RunningModeBinary {
		t.Fatalf("plane %q: running-mode %q not wired to a fixture yet - see https://trello.com/c/otriSswo",
			p.Name, p.Running)
	}

	return NewEnvironment(t, opts...)
}
