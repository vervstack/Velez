//go:build e2e_full

package e2e

import (
	"context"
	"net"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"
	"go.redsock.ru/toolbox"
	"go.vervstack.ru/matreshka/pkg/matreshka/resources"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients/state"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

// enableStatefullPgUnderDind brings up a TestEnvironment whose cluster
// postgres sidecar is published on a host-reachable DinD port
// (dindClusterPgPort) and whose ClusterPgDsn seam points the in-process
// host app's root DSN at that address, then runs EnablePlugin(statefull_pg)
// to completion. It returns the ready environment and the sidecar's
// container/volume name, and registers unconditional cleanup of both.
//
// No longer needs t.Chdir: sqldb.RollMigration (the create_schema_and_migrate
// job) resolves goose migrations relative to the process's working directory
// by default, so this helper points it at the repo's migrations/ directory
// via WithMigrationsDir instead (t.Chdir refuses to run under a parallel test
// or a parallel ancestor, which is what blocked t.Parallel() on this
// helper's callers before).
//
// Still NOT safe to call concurrently from two different callers: the
// sidecar is always published on the single fixed dindClusterPgPort
// (dind_ports.go), so two callers racing collide on that host port
// ("requested port is already occupied", confirmed against real Docker).
// EnableStatefullSuite, ServiceLifecycleSuite and VervonomiconDeploySuite -
// the three callers - all stay non-t.Parallel() at their top-level Test_X
// function for exactly this reason.
func enableStatefullPgUnderDind(t *testing.T, plane Plane, containerSuffix string) (*TestEnvironment, string) {
	t.Helper()

	hostAddr, ok := sharedDind.Addr(dindClusterPgPort)
	require.True(t, ok, "dind did not publish the cluster-pg port %d", dindClusterPgPort)

	host, portStr, err := net.SplitHostPort(hostAddr)
	require.NoError(t, err)

	port, err := strconv.ParseUint(portStr, 10, 64)
	require.NoError(t, err)

	// Only Host+Port are read back out by getRootDsnJob; user/pwd/dbname
	// still come from the sidecar container's own env vars. Build a
	// well-formed, parseable DSN anyway.
	advertisePg := &resources.Postgres{
		Host:    host,
		Port:    port,
		User:    "postgres",
		SslMode: "disable",
	}

	migrationsDir := filepath.Join(repoRoot(t), "migrations")

	env := plane.NewEnvironment(t,
		WithContainerSuffix(containerSuffix),
		WithClusterPgDsn(advertisePg.ConnectionString()),
		WithMigrationsDir(migrationsDir))

	pgName := state.PgName(containerSuffix)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	// The DinD daemon is named after the branch and reused across runs
	// (commit 6a4b620d), so its named volumes outlive a single `go test`.
	// A cluster-state pg volume left by an earlier failed run carries a
	// different generated postgres password and fails create_schema_and_migrate
	// with "password authentication failed". Wipe the sidecar + volume both
	// before enabling and on cleanup.
	removeStatefullSidecar(dockerClient, pgName)
	t.Cleanup(func() { removeStatefullSidecar(dockerClient, pgName) })

	statefullReq := &velez_api.EnableStatefullCluster{
		IsExposePort: toolbox.ToPtr(true),
		ExposeToPort: toolbox.ToPtr(uint64(dindClusterPgPort)),
	}
	payload := &velez_api.EnablePlugin_Request_StatefullCluster{
		StatefullCluster: statefullReq,
	}
	req := &velez_api.EnablePlugin_Request{
		Plugin:  velez_api.VervPluginType_statefull_pg,
		Payload: payload,
	}

	resp, err := env.Custom.ControlPlaneApiImpl.EnablePlugin(t.Context(), req)
	require.NoError(t, err)

	var finalTask tasks_queries.VelezTask

	for task := range env.Custom.JobsEngine.Watch(t.Context(), resp.GetEntityId(), resp.GetAction()) {
		finalTask = task
	}

	require.Equal(t, tasks_queries.VelezTaskStatusDONE, finalTask.Status,
		"enable statefull task error: %s", finalTask.Error.String)

	return env, pgName
}

// removeStatefullSidecar force-removes the cluster-state postgres container
// and its named volume, ignoring "no such container/volume". Both share the
// pgName.
func removeStatefullSidecar(dockerClient client.APIClient, pgName string) {
	ctx := context.Background()
	removeOpts := container.RemoveOptions{Force: true}

	_ = dockerClient.ContainerRemove(ctx, pgName, removeOpts)
	_ = dockerClient.VolumeRemove(ctx, pgName, true)
}
