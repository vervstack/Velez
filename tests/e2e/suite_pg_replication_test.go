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

const (
	replicationSourceName = "e2e-replication-source"
	replicationTargetName = "e2e-replication-target"
)

type PgReplicationSuite struct {
	suite.Suite
	plane Plane
}

func (s *PgReplicationSuite) Test_ReplicatePgDatabase() {
	t := s.T()
	t.Parallel()
	env := s.plane.NewEnvironment(t)
	ctx := t.Context()
	source := createReplicationPg(t, env, replicationSourceName)
	target := createReplicationPg(t, env, replicationTargetName)
	runReplicationSql(t, env, source, "CREATE TABLE replicated_rows (id serial PRIMARY KEY, value text NOT NULL)")
	runReplicationSql(t, env, source, "INSERT INTO replicated_rows (value) VALUES ('first'), ('second'), ('third')")
	req := &velez_api.ReplicatePgDatabase_Request{SourceInstanceId: source.GetName(), TargetInstanceId: target.GetName()}
	response, err := env.PgaasClient().ReplicatePgDatabase(ctx, req)
	require.NoError(t, err)
	require.Equal(t, jobs.ReplicatePgDatabaseAction, response.GetAction())
	finished := env.AwaitTask(t, response.GetEntityId(), response.GetAction())
	require.Equal(t, velez_api.TaskStatus_DONE, finished.GetStatus(), finished.GetError())
	rowCount := runReplicationSql(t, env, target, "SELECT count(*) FROM replicated_rows")
	require.Contains(t, rowCount, "3")
	sequenceValue := runReplicationSql(t, env, target, "SELECT last_value FROM replicated_rows_id_seq")
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

func runReplicationSql(t *testing.T, env *TestEnvironment, instance *velez_api.PgInstance, statement string) string {
	t.Helper()
	cli := env.DockerClient()
	args := filters.NewArgs(filters.Arg("label", labels.VervServiceLabel+"="+instance.GetName()))
	containers, err := cli.ContainerList(t.Context(), container.ListOptions{Filters: args})
	require.NoError(t, err)
	require.Len(t, containers, 1)
	execConfig := container.ExecOptions{
		Cmd:          []string{"psql", "-v", "ON_ERROR_STOP=1", "-U", instance.GetUsername(), "-d", instance.GetDbName(), "-c", statement},
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
