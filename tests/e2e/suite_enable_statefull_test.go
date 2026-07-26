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
}

func (s *EnableStatefullSuite) Test_EnableStatefullMode_UnsupportedPlugin_Fails() {
	t := s.T()
	t.Parallel()

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

	_, filename, _, _ := runtime.Caller(0)
	return filepath.Dir(filepath.Dir(filepath.Dir(filename)))
}
