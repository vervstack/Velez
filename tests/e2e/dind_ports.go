package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"

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
)

// dindPublishPorts is every container-side port the DinD daemon must
// publish: the matreshka bind plus the whole PortManager band.
func dindPublishPorts() []int {
	ports := make([]int, 0, dindPortBandEnd-dindMatreshkaPort+1)

	for p := dindMatreshkaPort; p <= dindPortBandEnd; p++ {
		ports = append(ports, p)
	}

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
