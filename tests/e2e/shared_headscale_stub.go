//go:build !e2e_full

package e2e

// stopSharedHeadscale is a no-op in the default/smoke build: shared_headscale.go
// (the real fixture, only used by suite_vpn_test.go) is tagged e2e_full, so
// there is nothing to stop here. See shared_headscale.go's stopSharedHeadscale
// for the e2e_full counterpart.
func stopSharedHeadscale() {}
