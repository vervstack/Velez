package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/cluster/env"
	"go.vervstack.ru/Velez/tests/dind"
)

// DinD port band.
//
// The DinD daemon (see main_test.go) publishes this whole container-side
// range to ephemeral bootstrap-host ports. Velez's PortManager is pinned to
// the band (see pinDindPorts) so every port Velez picks to expose a
// container lands on something the test process can reach through
// sharedDind.Addr. dindMatreshkaPort is carved out of the band for the
// shared matreshka fixture's fixed gRPC bind.
const (
	dindMatreshkaPort = 30000
	dindPortBandStart = 30001
	dindPortBandEnd   = 30019

	// dindClusterPgPort is carved out the same way dindMatreshkaPort is: the
	// DinD daemon publishes it (so the host process can reach the cluster-pg
	// sidecar), but it is deliberately kept OUT of the PortManager band
	// (dindAvailablePorts) so PortManager never hands it to another test's
	// smerd. The enable-statefull flow pins the sidecar's 5432 to this port
	// inside the DinD via EnableStatefullCluster.ExposeToPort.
	dindClusterPgPort = 30020
)

var (
	// sharedDind is the process-wide DinD handle, set once by runSuite
	// (main_test.go). Test helpers read it to translate a DinD-side
	// published port into the bootstrap-host address the process can dial.
	//nolint:gochecknoglobals // one disposable daemon per test binary
	sharedDind *dind.Env

	// dindSeedImages are images some code paths create a container from
	// without pulling first, so they must be pre-pulled into the fresh DinD
	// daemon. postgres:18 is pg_pattern.postgresImage (the cluster postgres
	// pattern).
	//nolint:gochecknoglobals // fixed suite input
	dindSeedImages = []string{"postgres:18"}

	// dindEnsureNetworks are docker networks the suite's smerds bind by name
	// that a real node already has but a fresh DinD does not: the "verv"
	// base network Velez no longer creates itself (env.StartNetwork is
	// disabled), plus "redsockru" (Test_StatelessMode_Loki).
	//nolint:gochecknoglobals // fixed suite input
	dindEnsureNetworks = []string{env.VervNetwork, "redsockru"}
)

// dindPublishPorts is every container-side port the DinD daemon must
// publish: the matreshka bind, the whole PortManager band, and the
// carved-out cluster-pg port (which is NOT in the band).
func dindPublishPorts() []int {
	bandLen := dindPortBandEnd - dindMatreshkaPort + 1

	ports := make([]int, 0, bandLen+1)

	for p := dindMatreshkaPort; p <= dindPortBandEnd; p++ {
		ports = append(ports, p)
	}

	ports = append(ports, dindClusterPgPort)

	return ports
}

// dindAvailablePorts is the PortManager band handed to Velez as
// Environment.AvailablePorts.
func dindAvailablePorts() []int {
	ports := make([]int, 0, dindPortBandEnd-dindPortBandStart+1)

	for p := dindPortBandStart; p <= dindPortBandEnd; p++ {
		ports = append(ports, p)
	}

	return ports
}

// dindHostAddr translates a DinD-side port Velez exposed a container on
// into the bootstrap-host address this process can dial it at.
func dindHostAddr(t *testing.T, exposedTo uint32) string {
	t.Helper()

	addr, ok := sharedDind.Addr(int(exposedTo))
	require.True(t, ok, "dind did not publish container port %d (outside the pinned band?)", exposedTo)

	return addr
}
