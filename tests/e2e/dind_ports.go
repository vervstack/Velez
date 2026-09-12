package e2e

import (
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

	// dindHeadscalePort is carved out the same way: the DinD daemon
	// publishes it so the in-process Velez app can reach the shared
	// headscale fixture (see shared_headscale.go) via
	// headscale.Connect(url, key). Kept OUT of the PortManager band so it
	// is never handed to a test's smerd. The fixture pins headscale's 8080
	// to this port inside the DinD.
	dindHeadscalePort = 30021
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
	// pattern); the headscale/tailscale images back the shared headscale
	// fixture and the VPN sidecar it exercises (see shared_headscale.go /
	// suite_vpn_test.go); alpine is copy_to_volume.go's
	// createLoaderContainerJob image, used by enable_registry's
	// htpasswd-loader step (see suite_enable_registry_test.go).
	//nolint:gochecknoglobals // fixed suite input
	dindSeedImages = []string{
		"postgres:18",
		"headscale/headscale:0.27.2-rc.1",
		"tailscale/tailscale:v1.90.8",
		"alpine",
	}

	// dindEnsureNetworks are docker networks the suite's smerds bind by name
	// that a real node already has but a fresh DinD does not: the "verv"
	// base network Velez no longer creates itself (env.StartNetwork is
	// disabled), plus "redsockru" (Test_StatelessMode_Loki).
	//nolint:gochecknoglobals // fixed suite input
	dindEnsureNetworks = []string{env.VervNetwork, "redsockru"}
)

// dindPublishPorts is every container-side port the DinD daemon must
// publish: the matreshka bind, the whole PortManager band, and the
// carved-out cluster-pg and headscale ports (which are NOT in the band).
func dindPublishPorts() []int {
	bandLen := dindPortBandEnd - dindMatreshkaPort + 1

	carvedOut := []int{dindClusterPgPort, dindHeadscalePort}

	ports := make([]int, 0, bandLen+len(carvedOut))

	for p := dindMatreshkaPort; p <= dindPortBandEnd; p++ {
		ports = append(ports, p)
	}

	ports = append(ports, carvedOut...)

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
