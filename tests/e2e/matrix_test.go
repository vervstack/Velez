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

const (
	ModeSingleNode PlaneMode = "single-node"
	ModeCluster    PlaneMode = "cluster"

	BackendDocker PlaneBackend = "docker"
)

// Plane names one cell of the single-node/cluster x runtime-backend test
// matrix, shared across every suite in this package. A suite adopts it by
// iterating Planes and building its fixture through Plane.NewEnvironment
// instead of calling NewEnvironment directly, so a new cell (a cluster
// fixture, a second runtime backend) lights up in every adopting suite at
// once instead of needing a per-suite matrix rewrite.
type Plane struct {
	Name    string
	Mode    PlaneMode
	Backend PlaneBackend
}

// Planes is the package-wide matrix of fixture cells. Only single-node/docker
// is wired to a real fixture today - see Plane.NewEnvironment.
var Planes = []Plane{
	{Name: "single-node/docker", Mode: ModeSingleNode, Backend: BackendDocker},

	// cluster/docker: needs WithMatreshka()+WithClusterPgDsn threaded through
	// per-suite before it can be a live cell here - see
	// docs/container_runtimes/roadmap.md. Not wired yet, deliberately.

	// A future non-docker Backend (podman, a dedicated instance) is a new
	// PlaneBackend value + a case in NewEnvironment once RuntimeResolver
	// grows a second implementation.
}

// NewEnvironment builds the TestEnvironment fixture matching p's Mode/Backend.
// Today both are single-case (single-node, docker), so this is a passthrough
// to package-level NewEnvironment. An unimplemented Mode/Backend combination
// fails loudly rather than silently building the wrong fixture, so adding a
// matrix row without wiring its fixture here is impossible to miss.
func (p Plane) NewEnvironment(t *testing.T, opts ...TestEnvOpt) *TestEnvironment {
	t.Helper()

	if p.Mode != ModeSingleNode {
		t.Fatalf("plane %q: mode %q not wired to a fixture yet", p.Name, p.Mode)
	}

	if p.Backend != BackendDocker {
		t.Fatalf("plane %q: backend %q not wired to a fixture yet", p.Name, p.Backend)
	}

	return NewEnvironment(t, opts...)
}
