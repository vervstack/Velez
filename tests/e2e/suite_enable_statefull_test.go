//go:build e2e_full

package e2e

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.redsock.ru/toolbox"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

// enableStatefullTestSuffix is this suite's ContainerSuffix, fixed and
// distinct from production's default empty suffix. It guarantees, by
// construction, that the container/volume name this suite creates
// (state.PgName(enableStatefullTestSuffix)) can never collide with - or
// destroy via this suite's unconditional t.Cleanup - a real, unsuffixed
// state.PgName("") instance a developer already has running on the same
// Docker host.
const (
	enableStatefullTestSuffix = "e2e-enable-statefull"
)

// EnableStatefullSuite exercises the "enable_statefull_mode" job through the
// real EnablePlugin RPC (internal/transport/control_plane_api_impl/enable_plugin.go),
// which enqueues onto the jobs engine and returns immediately with the
// task's (entityID, action) ref - it no longer blocks on Watch itself. The
// job's own 8-step orchestration already has unit-level coverage against
// fakes (internal/jobs/enable_statefull_test.go), whose own comments call
// out that the success path - create_schema_and_migrate/create_pg_user
// reaching a real Postgres - has no coverage anywhere but tests/e2e. This
// suite is that coverage; enableStatefullPgUnderDind does the waiting (via
// JobsEngine.Watch, the same way a real client would through
// TasksApi.WatchTask) before this suite asserts on the job's effects.
//
// The in-process host app reaches the cluster postgres sidecar (which runs
// inside the DinD daemon) through the ClusterPgDsn advertise-address seam:
// enableStatefullPgUnderDind pins the sidecar's 5432 to dindClusterPgPort
// (published by the DinD) and points ClusterPgDsn at the bootstrap-host
// address for that port.
//
// The postgres container/volume this job creates are named with this
// suite's own fixed enableStatefullTestSuffix (via WithContainerSuffix), not
// the bare production name - so this suite can never collide with, or
// force-remove, a real cluster-state postgres instance already running on
// the same Docker host. Only one instance of this suite can still run at a
// time on a given Docker host; the happy-path test isn't marked
// t.Parallel() and cleans up its own suffixed container/volume
// unconditionally.
type EnableStatefullSuite struct {
	suite.Suite

	plane Plane
}

func (s *EnableStatefullSuite) Test_EnableStatefullMode_HappyPath() {
	t := s.T()

	env, pgName := enableStatefullPgUnderDind(t, s.plane, enableStatefullTestSuffix)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	inspect, err := dockerClient.ContainerInspect(t.Context(), pgName)
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
		ServiceName: toolbox.ToPtr(pgName),
	}

	deploymentsResp, err := env.Custom.ServiceApiImpl.ListDeployments(t.Context(), listDeploymentsReq)
	require.NoError(t, err)
	require.EqualValues(t, 1, deploymentsResp.GetTotal())
	require.Len(t, deploymentsResp.GetDeployments(), 1)
	require.Equal(t, velez_api.DeploymentStatus_RUNNING, deploymentsResp.GetDeployments()[0].GetStatus())
}

func (s *EnableStatefullSuite) Test_EnableStatefullMode_UnsupportedPlugin_Fails() {
	t := s.T()

	env := s.plane.NewEnvironment(t)

	req := &velez_api.EnablePlugin_Request{
		Plugin: velez_api.VervPluginType_headscale,
	}

	_, err := env.Custom.ControlPlaneApiImpl.EnablePlugin(t.Context(), req)
	require.Error(t, err)
}

func Test_EnableStatefull(t *testing.T) {
	RunPlaneSuite(t, Planes, func(plane Plane) suite.TestingSuite {
		return &EnableStatefullSuite{plane: plane}
	})
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
