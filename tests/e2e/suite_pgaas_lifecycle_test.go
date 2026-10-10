//go:build e2e_full

package e2e

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

const (
	// pgaasLifecycleInstanceName is a fixed, suite-unique name. Unlike
	// jobs.RegistryServiceName (a node runs only one registry), pgaas has no
	// such singleton constraint - an instance name is caller-chosen and this
	// one doesn't collide with any other suite's fixed names, so this suite
	// is safe to run t.Parallel() against its siblings; it force-removes its
	// own container/volume before and after regardless.
	pgaasLifecycleInstanceName = "e2e-pgaas-lifecycle"

	// pgaasLifecycleServiceName is the name pgaas gives the service, container,
	// volume prefix and create_smerd task entity: the caller's name behind
	// labels.PgaasNamePrefix.
	pgaasLifecycleServiceName = labels.PgaasNamePrefix + pgaasLifecycleInstanceName
)

// PgaasLifecycleSuite exercises Postgres-as-a-Service end to end in plain
// single-node/local_storage mode (no WithMatreshka) - the mode the "instance
// is always called velez no matter the name I input" bug report was about.
// Before the fix, single-node ListPgInstances joined every pg_instances row
// onto service id 0 and resolved it to the synthetic "velez" service; the
// generated password and the row itself also lived only in memory and were
// lost on a Velez restart. The fix stamps VERV_SERVICE + velez.pgaas labels
// at create, derives a stable per-name int64 id, and reads the row + password
// back from the running container. This suite asserts the chosen name
// round-trips through list, that credentials resolve, and that drop removes
// the instance.
type PgaasLifecycleSuite struct {
	suite.Suite

	plane Plane
}

func (s *PgaasLifecycleSuite) Test_PgaasLifecycle_HappyPath() {
	t := s.T()
	t.Parallel()

	ctx := t.Context()

	env := s.plane.NewEnvironment(t)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	removePgaasInstance(dockerClient)
	t.Cleanup(func() { removePgaasInstance(dockerClient) })

	createReq := &velez_api.CreatePgInstance_Request{
		Name: pgaasLifecycleInstanceName,
	}

	createResp, err := env.Custom.PgaasApiImpl.CreatePgInstance(ctx, createReq)
	require.NoError(t, err)
	require.Equal(t, pgaasLifecycleInstanceName, createResp.GetEntityId())
	require.Equal(t, jobs.CreatePgInstanceAction, createResp.GetAction())

	var createTask tasks_queries.VelezTask

	for task := range env.Custom.JobsEngine.Watch(ctx, createResp.GetEntityId(), createResp.GetAction()) {
		createTask = task
	}

	require.Equal(t, tasks_queries.VelezTaskStatusDONE, createTask.Status,
		"create_pg_instance task error: %s", createTask.Error.String)

	instance := findPgInstance(t, env, pgaasLifecycleServiceName)
	require.NotNil(t, instance, "expected pg instance %q in ListPgInstances", pgaasLifecycleServiceName)
	require.Equal(t, pgaasLifecycleServiceName, instance.GetName())
	require.NotEqual(t, "velez", instance.GetName())
	require.Equal(t, "running", instance.GetStatus())
	require.Equal(t, pgaasLifecycleInstanceName, listedServiceDisplayName(t, env, pgaasLifecycleServiceName))

	// Credentials resolve through the container-env read-through path - there
	// is no velez.secrets table in single-node mode.
	credsReq := &velez_api.GetPgInstanceCredentials_Request{Name: pgaasLifecycleServiceName}

	credsResp, err := env.Custom.PgaasApiImpl.GetPgInstanceCredentials(ctx, credsReq)
	require.NoError(t, err)
	require.NotEmpty(t, credsResp.GetPassword())
	require.NotEmpty(t, credsResp.GetDsn())
	require.Equal(t, instance.GetDbName(), credsResp.GetDbName())
	require.Equal(t, instance.GetUsername(), credsResp.GetUsername())

	dropReq := &velez_api.DropPgInstance_Request{Name: pgaasLifecycleServiceName}

	dropResp, err := env.Custom.PgaasApiImpl.DropPgInstance(ctx, dropReq)
	require.NoError(t, err)
	require.Equal(t, pgaasLifecycleServiceName, dropResp.GetEntityId())
	require.Equal(t, jobs.DropPgInstanceAction, dropResp.GetAction())

	var dropTask tasks_queries.VelezTask

	for task := range env.Custom.JobsEngine.Watch(ctx, dropResp.GetEntityId(), dropResp.GetAction()) {
		dropTask = task
	}

	require.Equal(t, tasks_queries.VelezTaskStatusDONE, dropTask.Status,
		"drop_pg_instance task error: %s", dropTask.Error.String)

	dropped := findPgInstance(t, env, pgaasLifecycleServiceName)
	require.Nil(t, dropped, "pg instance still listed after drop")
}

func Test_PgaasLifecycle(t *testing.T) {
	t.Parallel()
	RunPlaneSuite(t, Planes, func(plane Plane) suite.TestingSuite {
		return &PgaasLifecycleSuite{plane: plane}
	})
}

// findPgInstance returns the listed pg instance with the given name, or nil.
func findPgInstance(t *testing.T, env *TestEnvironment, name string) *velez_api.PgInstance {
	t.Helper()

	listReq := &velez_api.ListPgInstances_Request{}

	listResp, err := env.Custom.PgaasApiImpl.ListPgInstances(t.Context(), listReq)
	require.NoError(t, err)

	for _, pg := range listResp.GetInstances() {
		if pg.GetName() == name {
			return pg
		}
	}

	return nil
}

// removePgaasInstance force-removes the fixed-name pg instance container and
// its per-instance data volume (pgaas.pgVolumeName is "<service name>-data"),
// ignoring "no such container/volume".
func removePgaasInstance(dockerClient client.APIClient) {
	ctx := context.Background()
	removeOpts := container.RemoveOptions{Force: true}

	_ = dockerClient.ContainerRemove(ctx, pgaasLifecycleServiceName, removeOpts)
	_ = dockerClient.VolumeRemove(ctx, pgaasLifecycleServiceName+"-data", true)
}
