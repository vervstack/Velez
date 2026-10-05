//go:build e2e_full

package e2e

import (
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

const (
	clusterPgParallelEnv = "VELEZ_E2E_CLUSTER_PG_PARALLEL"

	clusterPgLeaseTimeout = 20 * time.Minute

	// clusterPgDefaultParallel caps concurrent cluster sidecars when the env
	// var is unset: at 12 concurrent real postgres sidecars plus the parallel
	// non-cluster tests the shared DinD daemon intermittently times out
	// create_smerd / deployment-sync waits; 4 was stable.
	clusterPgDefaultParallel = 4
)

// clusterPgPool hands out dedicated host ports for the cluster-pg sidecar, so
// every cluster-mode test gets its own sidecar instead of racing for one fixed
// port. Its size doubles as the cap on concurrently running cluster sidecars
// (each one is a real postgres): VELEZ_E2E_CLUSTER_PG_PARALLEL sets it, up to
// dindClusterPgPoolSize.
//
//nolint:gochecknoglobals // one pool per test binary
var (
	clusterPgPoolOnce sync.Once
	clusterPgPool     chan int

	suffixLocksMu sync.Mutex
	suffixLocks   = map[string]chan struct{}{}
)

func clusterPgPoolSize() int {
	raw := os.Getenv(clusterPgParallelEnv)
	if raw == "" {
		return clusterPgDefaultParallel
	}

	limit, err := strconv.Atoi(raw)
	if err != nil || limit < 1 || limit > dindClusterPgPoolSize {
		return clusterPgDefaultParallel
	}

	return limit
}

func initClusterPgPool() {
	size := clusterPgPoolSize()

	clusterPgPool = make(chan int, size)

	for i := range size {
		clusterPgPool <- dindClusterPgPoolStart + i
	}
}

// acquireClusterPgPort leases a free cluster-pg host port for the test and
// returns it to the pool in t.Cleanup. It blocks while every port is leased.
// The lease is taken first in the cleanup order of the caller, so it is
// released last: after the sidecar, which owns the port, is already removed.
func acquireClusterPgPort(t *testing.T) int {
	t.Helper()

	clusterPgPoolOnce.Do(initClusterPgPool)

	timeout := time.NewTimer(clusterPgLeaseTimeout)
	defer timeout.Stop()

	var port int

	select {
	case port = <-clusterPgPool:
	case <-timeout.C:
		t.Fatalf("no free cluster-pg port within %s", clusterPgLeaseTimeout)
	case <-t.Context().Done():
		t.Fatalf("test finished while waiting for a cluster-pg port")
	}

	t.Cleanup(func() { clusterPgPool <- port })

	return port
}

// lockContainerSuffix serializes the tests sharing a container suffix: the
// sidecar, its volume and the environment's container names derive from it, so
// two concurrent holders of one suffix would collide. The lock is released in
// t.Cleanup.
func lockContainerSuffix(t *testing.T, suffix string) {
	t.Helper()

	suffixLocksMu.Lock()

	lock, isKnown := suffixLocks[suffix]
	if !isKnown {
		lock = make(chan struct{}, 1)
		suffixLocks[suffix] = lock
	}

	suffixLocksMu.Unlock()

	timeout := time.NewTimer(clusterPgLeaseTimeout)
	defer timeout.Stop()

	select {
	case lock <- struct{}{}:
	case <-timeout.C:
		t.Fatalf("container suffix %q stayed locked for %s", suffix, clusterPgLeaseTimeout)
	case <-t.Context().Done():
		t.Fatalf("test finished while waiting for container suffix %q", suffix)
	}

	t.Cleanup(func() { <-lock })
}
