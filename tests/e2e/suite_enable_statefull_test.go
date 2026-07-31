package e2e

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.redsock.ru/toolbox"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients/state"
)

// EnableStatefullSuite exercises the "enable_statefull_mode" job through the
// real EnablePlugin RPC (internal/transport/control_plane_api_impl/enable_plugin.go),
// which enqueues onto the jobs engine and blocks on Watch. The job's own
// 8-step orchestration already has unit-level coverage against fakes
// (internal/jobs/enable_statefull_test.go), whose own comments call out that
// the success path - create_schema_and_migrate/create_pg_user reaching a
// real Postgres - has no coverage anywhere but tests/e2e. This suite is that
// coverage.
//
// The postgres container/volume this job creates use the fixed name
// state.PgName (no per-test suffix - that's existing production behavior,
// not something this test controls), so only one instance of this suite can
// run at a time on a given Docker host; the happy-path test isn't marked
// t.Parallel() and cleans up its container/volume unconditionally.
type EnableStatefullSuite struct {
	suite.Suite
}

func (s *EnableStatefullSuite) Test_EnableStatefullMode_HappyPath() {
	t := s.T()

	env := NewEnvironment(t)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	t.Cleanup(func() {
		ctx := context.Background()
		removeOpts := container.RemoveOptions{Force: true}

		_ = dockerClient.ContainerRemove(ctx, state.PgName, removeOpts)
		_ = dockerClient.VolumeRemove(ctx, state.PgName, true)
	})

	// sqldb.RollMigration (called by the create_schema_and_migrate job) rolls
	// goose migrations from the hardcoded relative path "./migrations" - a
	// pre-existing constraint of that function (see enable_statefull.go's
	// createSchemaAndMigrateJob comment), not something this test controls.
	// `go test` runs with cwd = the package directory, so it must be pointed
	// at the repo root for the duration of this test.
	t.Chdir(repoRoot(t))

	statefullReq := &velez_api.EnableStatefullCluster{
		IsExposePort: toolbox.ToPtr(true),
	}
	payload := &velez_api.EnablePlugin_Request_StatefullCluster{
		StatefullCluster: statefullReq,
	}
	req := &velez_api.EnablePlugin_Request{
		Plugin:  velez_api.VervPluginType_statefull_pg,
		Payload: payload,
	}

	_, err := env.Custom.ControlPlaneApiImpl.EnablePlugin(t.Context(), req)
	require.NoError(t, err)

	inspect, err := dockerClient.ContainerInspect(t.Context(), state.PgName)
	require.NoError(t, err)
	require.NotNil(t, inspect.State)
	require.True(t, inspect.State.Running)

	localState := env.Custom.NodeClients.LocalStateManager().Get()
	require.NotEmpty(t, localState.ClusterState.PgRootDsn)
	require.NotEmpty(t, localState.ClusterState.PgNodeDsn)

	// registerPluginJob (internal/jobs/enable_statefull.go) is the fix for the
	// bug where ListPlugins always reported statefull_pg as VervPlugin_unknown
	// because deployment_specifications.service_id was never populated. Assert
	// the full chain - CreateSpecification's new ServiceID column through to
	// the ListPlugins join - actually reports the plugin as running now.
	listPluginsResp, err := env.Custom.ControlPlaneApiImpl.ListPlugins(t.Context(), &velez_api.ListPlugins_Request{})
	require.NoError(t, err)

	var statefullPgPlugin *velez_api.Plugin

	for _, plugin := range listPluginsResp.GetPlugins() {
		if plugin.GetType() == velez_api.VervPluginType_statefull_pg {
			statefullPgPlugin = plugin

			break
		}
	}

	require.NotNil(t, statefullPgPlugin, "expected statefull_pg plugin in ListPlugins response")
	require.Equal(t, velez_api.VervPlugin_running, statefullPgPlugin.GetState())

	// registerPluginJob also leaves behind a real velez.deployments row
	// (spec_id -> deployment_specifications.service_id, never a service_id
	// column on deployments itself). ListDeployments used to select
	// "service_id" straight off velez.deployments, which errored with
	// "column \"service_id\" does not exist" as soon as this path was hit
	// against real Postgres - internal/storage/postgres/deployments.go now
	// joins deployment_specifications to resolve it. Exercise the ServiceName
	// filter here since that's the clause that touches the joined column.
	//
	// Goes through the real RPC (env.Custom.ServiceApiImpl.ListDeployments)
	// rather than querying env.Custom.Services.StorageContainer() directly,
	// as it used to: VervService (internal/service/service_manager/verv_services/service.go)
	// used to resolve and cache storageContainer.Deployments() once at
	// construction time, before this test's EnablePlugin call swaps the
	// container over to the real Postgres backend, so the RPC path always
	// saw zero rows against the stale pre-swap storage. VervService now holds
	// the storage.Storage container itself and resolves Deployments() fresh
	// on every call, so the RPC path reflects the swap correctly - this
	// assertion going through it is what proves that fix.
	listDeploymentsReq := &velez_api.ListDeployments_Request{
		ServiceName: toolbox.ToPtr(state.PgName),
	}

	deploymentsResp, err := env.Custom.ServiceApiImpl.ListDeployments(t.Context(), listDeploymentsReq)
	require.NoError(t, err)
	require.EqualValues(t, 1, deploymentsResp.GetTotal())
	require.Len(t, deploymentsResp.GetDeployments(), 1)
	require.Equal(t, velez_api.DeploymentStatus_RUNNING, deploymentsResp.GetDeployments()[0].GetStatus())
}

func (s *EnableStatefullSuite) Test_EnableStatefullMode_UnsupportedPlugin_Fails() {
	t := s.T()

	env := NewEnvironment(t)

	req := &velez_api.EnablePlugin_Request{
		Plugin: velez_api.VervPluginType_headscale,
	}

	_, err := env.Custom.ControlPlaneApiImpl.EnablePlugin(t.Context(), req)
	require.Error(t, err)
}

func Test_EnableStatefull(t *testing.T) {
	suite.Run(t, new(EnableStatefullSuite))
}

// repoRoot resolves the repository root (the directory containing
// migrations/) the same way helper_environment.go's testsDir does: from this
// file's own location, two directories up from tests/e2e.
func repoRoot(t *testing.T) string {
	t.Helper()

	pc, filename, _, _ := runtime.Caller(0)

	_ = pc

	return filepath.Dir(filepath.Dir(filepath.Dir(filename)))
}
