package e2e

import (
	"go.vervstack.ru/Velez/internal/cluster/env"
	"go.vervstack.ru/Velez/tests/dind"
)

// DinD port band.
//
// The DinD daemon (see main_test.go) publishes this whole container-side
// range to ephemeral bootstrap-host ports. Velez's PortManager is pinned to
// the allocator range (see pinDindPorts) so every port Velez picks to expose
// a container lands on something the test process can reach through
// sharedDind.Addr.
//
// The band is laid out as:
//
//	dindMatreshkaPort                     shared matreshka fixture's gRPC bind
//	dindPortBandStart..dindAllocatorEnd   PortManager allocator range
//	dindReservedStart..dindPortBandEnd    ports tests bind themselves
//	dindHeadscalePort                     fixture port, past the band
//	dindClusterPgPoolStart..+PoolSize-1   cluster-pg sidecar ports, one leased per test
//
// Velez never releases a port it handed out to a smerd while the process
// lives, so the allocator range must cover every allocation of one full
// parallel run, not just the peak of concurrently running tests - an
// exhausted pool makes the PortManager re-hand a port still bound inside the
// DinD ("port is already allocated").
//
// A test that binds a host port itself (foreign containers registered into
// Velez) must take it from the reserved block as dindPortBandEnd-N
// (N < dindReservedPortCount), never from the allocator range, or the
// PortManager may hand the same port to a parallel test.
const (
	dindMatreshkaPort = 30000
	dindPortBandStart = 30001
	dindPortBandEnd   = 30200

	dindReservedPortCount = 10
	dindAllocatorEnd      = dindPortBandEnd - dindReservedPortCount

	// dindClusterPgPoolStart..dindClusterPgPoolStart+dindClusterPgPoolSize-1
	// are carved out the same way dindMatreshkaPort is: the DinD daemon
	// publishes them (so the host process can reach a cluster-pg sidecar), but
	// they are deliberately kept OUT of the PortManager range
	// (dindAvailablePorts) so PortManager never hands one to another test's
	// smerd. Each cluster-mode test leases one (acquireClusterPgPort) and the
	// enable-statefull flow pins the sidecar's 5432 to it inside the DinD via
	// EnableStatefullCluster.ExposeToPort.
	dindClusterPgPoolStart = dindPortBandEnd + 3
	dindClusterPgPoolSize  = 12

	// dindHeadscalePort is carved out the same way: the DinD daemon
	// publishes it so the in-process Velez app can reach the shared
	// headscale fixture (see shared_headscale.go) via
	// headscale.Connect(url, key). Kept OUT of the PortManager range so it
	// is never handed to a test's smerd. The fixture pins headscale's 8080
	// to this port inside the DinD.
	dindHeadscalePort = dindPortBandEnd + 2
)

var (
	// sharedDind is the process-wide DinD handle, set once by runSuite
	// (main_test.go). Test helpers read it to translate a DinD-side
	// published port into the bootstrap-host address the process can dial.
	//nolint:gochecknoglobals // one disposable daemon per test binary
	sharedDind *dind.Env

	// sharedTargetDind is the second daemon, the transfer target's; brought up
	// on first use (see helper_transfer_target_test.go) and torn down by
	// runSuite.
	//nolint:gochecknoglobals // one disposable daemon per test binary
	sharedTargetDind *dind.Env

	// dindSeedImages are the images the suite and Velez use inside the DinD.
	// The DinD never pulls from the web: each is loaded from the host image
	// cache (see dind.Env.Seed). Floating names the product or tests request
	// verbatim (alpine, :latest, registry:2) are pinned to a concrete version
	// and tagged back as aliases. Some code paths create a container from
	// them without pulling first. postgres:18 is pg_pattern.postgresImage (the cluster postgres
	// pattern); the headscale/tailscale images back the shared headscale
	// fixture and the VPN sidecar it exercises (see shared_headscale.go /
	// suite_vpn_test.go); alpine is copy_to_volume.go's
	// createLoaderContainerJob image, used by enable_registry's
	// htpasswd-loader step (see suite_enable_registry_test.go). The binfmt
	// installer is the node-side image of set_runner_buildkit (see
	// suite_runner_buildkit_test.go), seeded so that suite skips a cold pull.
	// The rest are every other image the suite or Velez pulls inside the
	// daemon while tests run; matreshka is pinned to the go.mod module version.
	//nolint:gochecknoglobals // fixed suite input
	dindSeedImages = []dind.SeedImage{
		{Ref: "postgres:18"},
		{Ref: "postgres:16"},
		{Ref: "headscale/headscale:0.27.2-rc.1"},
		{Ref: "tailscale/tailscale:v1.90.8"},
		{Ref: "alpine:3.24.2", Aliases: []string{"alpine", "alpine:3"}},
		{Ref: "nginx:alpine"},
		{Ref: "godverv/hello_world:v0.0.14"},
		{Ref: "vervstack/hello_world:v0.0.15"},
		{Ref: "vervstack/matreshka:v1.0.102"},
		{Ref: "grafana/loki:main-bc418c4"},
		{Ref: "gitlab/gitlab-runner:v19.5.0", Aliases: []string{"gitlab/gitlab-runner:latest"}},
		{Ref: "registry:2.8.3", Aliases: []string{"registry:2"}},
		{Ref: "joxit/docker-registry-ui:2.6.0", Aliases: []string{"joxit/docker-registry-ui:latest"}},
		{Ref: "dxflrs/garage:v2.4.0"},
		{Ref: "khairul169/garage-webui:1.1.0"},
		{Ref: "moby/buildkit:v0.34.0"},
		{Ref: "tonistiigi/binfmt:qemu-v10.2.3-68"},
	}

	// dindEnsureNetworks are docker networks the suite's smerds bind by name
	// that a real node already has but a fresh DinD does not: the "verv"
	// base network Velez no longer creates itself (env.StartNetwork is
	// disabled), plus "redsockru" (Test_StatelessMode_Loki).
	//nolint:gochecknoglobals // fixed suite input
	dindEnsureNetworks = []string{env.VervNetwork, "redsockru"}
)

// dindPublishPorts is every container-side port the DinD daemon must
// publish: the matreshka bind, the whole band (allocator range plus the
// reserved test block), and the fixture ports past the band.
func dindPublishPorts() []int {
	ports := make([]int, 0, dindPortBandEnd-dindMatreshkaPort+1+1+dindClusterPgPoolSize)

	for p := dindMatreshkaPort; p <= dindPortBandEnd; p++ {
		ports = append(ports, p)
	}

	ports = append(ports, dindHeadscalePort)

	for i := range dindClusterPgPoolSize {
		ports = append(ports, dindClusterPgPoolStart+i)
	}

	return ports
}

// dindAvailablePorts is the PortManager range handed to Velez as
// Environment.AvailablePorts: the band minus the matreshka port and the
// reserved block tests bind themselves.
func dindAvailablePorts() []int {
	ports := make([]int, 0, dindAllocatorEnd-dindPortBandStart+1)

	for p := dindPortBandStart; p <= dindAllocatorEnd; p++ {
		ports = append(ports, p)
	}

	return ports
}
