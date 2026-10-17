package e2e

import (
	"os"
	"slices"

	"go.vervstack.ru/Velez/internal/cluster/env"
	"go.vervstack.ru/Velez/tests/dind"
)

// DinD port plan.
//
// Every port below is bound inside the DinD daemon and reached from this
// process through the group's SOCKS5 proxy at the verv network's gateway
// (dindGatewayAddr); the DinD publishes nothing else besides its Docker API and
// the proxy.
//
//	dindMatreshkaPort                     shared matreshka fixture's gRPC bind
//	dindPortBandStart..dindAllocatorEnd   per-Velez slot slices (AvailablePorts)
//	dindReservedStart..dindPortBandEnd    ports tests bind themselves
//	dindHeadscalePort                     shared headscale fixture
//	dindClusterPgPoolStart..+PoolSize-1   cluster-pg sidecar ports, one leased per test
//
// A test that binds a port itself (foreign containers registered into Velez)
// must take it from the reserved block as dindPortBandEnd-N
// (N < dindReservedPortCount), never from a slot slice, or a Velez may hand
// the same port to a parallel test.
const (
	envVelezImage = "VELEZ_E2E_VELEZ_IMAGE"

	dindMatreshkaPort = 30000
	dindPortBandStart = 30001
)

var (
	// sharedDind is the process-wide DinD handle, set once by runSuite
	// (main_test.go). Test helpers read it to translate a DinD-side
	// published port into the bootstrap-host address the process can dial.
	//nolint:gochecknoglobals // one disposable daemon per test binary
	sharedDind *dind.Env

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

// seedImages is dindSeedImages plus the locally built Velez image when
// e2egroups passes its ref (VELEZ_E2E_VELEZ_IMAGE).
func seedImages() []dind.SeedImage {
	images := slices.Clone(dindSeedImages)

	velezImage := os.Getenv(envVelezImage)
	if velezImage == "" {
		return images
	}

	local := dind.SeedImage{Ref: velezImage, IsLocalOnly: true}

	return append(images, local)
}
