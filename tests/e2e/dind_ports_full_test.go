//go:build e2e_full

package e2e

const (
	dindPortBandEnd = 30200

	dindReservedPortCount = 10
	dindAllocatorEnd      = dindPortBandEnd - dindReservedPortCount

	// dindClusterPgPoolStart..dindClusterPgPoolStart+dindClusterPgPoolSize-1
	// are carved out the same way dindMatreshkaPort is: kept OUT of the slot
	// slices so no Velez hands one to another test's smerd. Each cluster-mode
	// test leases one (acquireClusterPgPort) and the enable-statefull flow pins
	// the sidecar's 5432 to it inside the DinD via
	// EnableStatefullCluster.ExposeToPort.
	dindClusterPgPoolStart = dindPortBandEnd + 3
	dindClusterPgPoolSize  = 12

	// dindHeadscalePort is carved out the same way: Velez reaches the shared
	// headscale fixture (see shared_headscale.go) via
	// headscale.Connect(url, key). Kept OUT of the slot slices so it is
	// never handed to a test's smerd. The fixture pins headscale's 8080 to
	// this port inside the DinD.
	dindHeadscalePort = dindPortBandEnd + 2
)
