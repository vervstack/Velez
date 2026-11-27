//go:build e2e_full

package e2e

import (
	"bytes"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/jobs"
)

type PgReplicationSuite struct {
	suite.Suite
	plane Plane
}

// Instance names carry a per-plane suffix: both planes share one Docker daemon,
// so fixed names would make the parallel planes fight over the same containers.
// Instance names are limited to 32 characters.
func (s *PgReplicationSuite) replicationSourceName() string {
	return "e2e-replication-src-" + s.planeSuffix()
}

func (s *PgReplicationSuite) replicationTargetName() string {
	return "e2e-replication-tgt-" + s.planeSuffix()
}

func (s *PgReplicationSuite) planeSuffix() string {
	if s.plane.Mode == ModeCluster {
		return "cl"
	}

	return "sn"
}

func (s *PgReplicationSuite) Test_ReplicatePgDatabase() {
	t := s.T()
	t.Parallel()
	env := s.plane.NewEnvironment(t)
	ctx := t.Context()
	source := createReplicationPg(t, env, s.replicationSourceName())
	target := createReplicationPg(t, env, s.replicationTargetName())
	runReplicationSql(t, env, source, source.GetDbName(), "CREATE TABLE replicated_rows (id serial PRIMARY KEY, value text NOT NULL)")
	runReplicationSql(t, env, source, source.GetDbName(), "INSERT INTO replicated_rows (value) VALUES ('first'), ('second'), ('third')")
	req := &velez_api.ReplicatePgDatabase_Request{SourceInstanceId: source.GetName(), TargetInstanceId: target.GetName()}
	response, err := env.PgaasClient().ReplicatePgDatabase(ctx, req)
	require.NoError(t, err)
	require.Equal(t, jobs.ReplicatePgDatabaseAction, response.GetAction())
	finished := env.AwaitTask(t, response.GetEntityId(), response.GetAction())
	require.Equal(t, velez_api.TaskStatus_DONE, finished.GetStatus(), finished.GetError())
	// The dump restores into a database named after the SOURCE database on the
	// target instance - the target's own database stays untouched.
	rowCount := runReplicationSql(t, env, target, source.GetDbName(), "SELECT count(*) FROM replicated_rows")
	require.Contains(t, rowCount, "3")
	sequenceValue := runReplicationSql(t, env, target, source.GetDbName(), "SELECT last_value FROM replicated_rows_id_seq")
	require.Contains(t, sequenceValue, "3")
	repeated, err := env.PgaasClient().ReplicatePgDatabase(ctx, req)
	require.NoError(t, err)
	failed := env.AwaitTask(t, repeated.GetEntityId(), repeated.GetAction())
	require.Equal(t, velez_api.TaskStatus_FAILED, failed.GetStatus())
	require.Contains(t, failed.GetError(), "target database already exists")
}

func createReplicationPg(t *testing.T, env *TestEnvironment, name string) *velez_api.PgInstance {
	t.Helper()
	request := &velez_api.CreatePgInstance_Request{Name: name}
	response, err := env.PgaasClient().CreatePgInstance(t.Context(), request)
	require.NoError(t, err)
	completed := env.AwaitTask(t, response.GetEntityId(), response.GetAction())
	require.Equal(t, velez_api.TaskStatus_DONE, completed.GetStatus(), completed.GetError())
	return findPgInstance(t, env, labels.PgaasNamePrefix+name)
}

// runReplicationSql execs psql inside the container backing instance, connecting
// as that instance's user to the database dbName.
func runReplicationSql(t *testing.T, env *TestEnvironment, instance *velez_api.PgInstance, dbName, statement string) string {
	t.Helper()
	cli := env.DockerClient()
	args := filters.NewArgs(filters.Arg("label", labels.VervServiceLabel+"="+instance.GetName()))
	containers, err := cli.ContainerList(t.Context(), container.ListOptions{Filters: args})
	require.NoError(t, err)
	require.Len(t, containers, 1)
	execConfig := container.ExecOptions{
		Cmd:          []string{"psql", "-v", "ON_ERROR_STOP=1", "-U", instance.GetUsername(), "-d", dbName, "-c", statement},
		AttachStdout: true,
		AttachStderr: true,
	}
	exec, err := cli.ContainerExecCreate(t.Context(), containers[0].ID, execConfig)
	require.NoError(t, err)
	attached, err := cli.ContainerExecAttach(t.Context(), exec.ID, container.ExecAttachOptions{})
	require.NoError(t, err)
	defer attached.Close()
	var stdout, stderr bytes.Buffer
	_, err = stdcopy.StdCopy(&stdout, &stderr, attached.Reader)
	require.NoError(t, err)
	inspection, err := cli.ContainerExecInspect(t.Context(), exec.ID)
	require.NoError(t, err)
	require.Zero(t, inspection.ExitCode, stderr.String())
	return stdout.String()
}

func Test_PgReplication(t *testing.T) {
	t.Parallel()
	RunPlaneSuite(t, Planes, func(plane Plane) suite.TestingSuite { return &PgReplicationSuite{plane: plane} })
}
