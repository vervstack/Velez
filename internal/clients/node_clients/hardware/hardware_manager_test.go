package hardware

import (
	"testing"
	"time"

	"go.vervstack.ru/Velez/internal/cluster/env/containerinfo"
)

// containerinfo.IsInContainer() memoizes a real filesystem check
// (/etc/hostname readable) for the process lifetime, so this test asserts
// GetHardware's reported value matches it directly rather than hard-coding
// an expectation - the same approach enable_statefull_test.go uses for the
// equivalent env.IsInContainer()-dependent branches.
func TestManager_GetHardware_IsRunningInContainer(t *testing.T) {
	m := New("")

	resp, err := m.GetHardware()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := containerinfo.IsInContainer()
	if resp.GetIsRunningInContainer() != want {
		t.Errorf("expected IsRunningInContainer to be %v, got %v", want, resp.GetIsRunningInContainer())
	}
}

// GetHardware re-runs ghw's CPU/Memory/Block discovery on every uncached call, which
// is too costly to do on demand for every dialog render - it's cached and only
// refreshed once hardwareCacheTTL has elapsed since the last computation.
func TestManager_GetHardware_CachesWithinTTL(t *testing.T) {
	m := New("")

	first, err := m.GetHardware()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	second, err := m.GetHardware()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if first != second {
		t.Errorf("expected GetHardware to return the cached response within the TTL, got distinct instances")
	}
}

// New's region argument is passed straight through into every GetHardware
// response, so callers wiring the manager from config.Environment.NodeRegion
// (see internal/clients/node_clients/node_clients.go and
// internal/transport/velez_api_impl/impl.go) can rely on GetHardware
// reporting the node's configured region.
func TestManager_GetHardware_NodeRegionPassesThrough(t *testing.T) {
	const region = "eu-west-1"

	m := New(region)

	resp, err := m.GetHardware()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.GetNodeRegion() != region {
		t.Errorf("expected NodeRegion to be %q, got %q", region, resp.GetNodeRegion())
	}
}

func TestManager_GetHardware_RefreshesAfterTTL(t *testing.T) {
	original := hardwareCacheTTL

	hardwareCacheTTL = time.Millisecond

	defer func() { hardwareCacheTTL = original }()

	m := New("")

	first, err := m.GetHardware()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	time.Sleep(2 * time.Millisecond)

	second, err := m.GetHardware()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if first == second {
		t.Errorf("expected GetHardware to refresh after the TTL elapsed, got the same cached instance")
	}
}
