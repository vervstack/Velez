//go:build e2e_full

package e2e

import (
	"context"
	"net"
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
// The caller MUST NOT be a t.Parallel() test: sqldb.RollMigration (the
// create_schema_and_migrate job) rolls goose migrations from the hardcoded
// relative "./migrations", so this helper t.Chdir's to the repo root for the
// duration of the test.
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

	t.Chdir(repoRoot(t))

	env := plane.NewEnvironment(t,
		WithContainerSuffix(containerSuffix),
		WithClusterPgDsn(advertisePg.ConnectionString()))

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
