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
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

const (
	// pgaasLifecycleInstanceName is a fixed, unsuffixed name: this suite is
	// not t.Parallel() and force-removes its own container/volume before and
	// after, the same trade-off enableRegistryUnderDind makes for
	// jobs.RegistryServiceName.
	pgaasLifecycleInstanceName = "e2e-pgaas-lifecycle"
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
}

func (s *PgaasLifecycleSuite) Test_PgaasLifecycle_HappyPath() {
	t := s.T()
	ctx := t.Context()

	env := Planes[0].NewEnvironment(t)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	removePgaasInstance(dockerClient)
	t.Cleanup(func() { removePgaasInstance(dockerClient) })

	createReq := &velez_api.CreatePgInstance_Request{
		Name: pgaasLifecycleInstanceName,
	}

	_, err := env.Custom.PgaasApiImpl.CreatePgInstance(ctx, createReq)
	require.NoError(t, err)

	// CreatePgInstance only writes a SCHEDULED_DEPLOYMENT row; the deploy
	// watcher dispatches the create_smerd task that actually launches the
	// container (see enableRegistryUnderDind's doc comment). Environment is
	// empty (the create dialog's default), so SmerdEntityID is the bare name.
	entityID := jobs.SmerdEntityID("", pgaasLifecycleInstanceName)

	var deployTask tasks_queries.VelezTask

	for task := range env.Custom.JobsEngine.Watch(ctx, entityID, jobs.CreateSmerdAction) {
		deployTask = task
	}

	require.Equal(t, tasks_queries.VelezTaskStatusDONE, deployTask.Status,
		"create_smerd task for pg instance error: %s", deployTask.Error.String)

	instance := findPgInstance(t, env, pgaasLifecycleInstanceName)
	require.NotNil(t, instance, "expected pg instance %q in ListPgInstances", pgaasLifecycleInstanceName)
	require.Equal(t, pgaasLifecycleInstanceName, instance.GetName())
	require.NotEqual(t, "velez", instance.GetName())
	require.Equal(t, "running", instance.GetStatus())

	// Credentials resolve through the container-env read-through path - there
	// is no velez.secrets table in single-node mode.
	credsReq := &velez_api.GetPgInstanceCredentials_Request{Name: pgaasLifecycleInstanceName}

	credsResp, err := env.Custom.PgaasApiImpl.GetPgInstanceCredentials(ctx, credsReq)
	require.NoError(t, err)
	require.NotEmpty(t, credsResp.GetPassword())
	require.NotEmpty(t, credsResp.GetDsn())
	require.Equal(t, instance.GetDbName(), credsResp.GetDbName())
	require.Equal(t, instance.GetUsername(), credsResp.GetUsername())

	dropReq := &velez_api.DropPgInstance_Request{Name: pgaasLifecycleInstanceName}

	_, err = env.Custom.PgaasApiImpl.DropPgInstance(ctx, dropReq)
	require.NoError(t, err)

	dropped := findPgInstance(t, env, pgaasLifecycleInstanceName)
	require.Nil(t, dropped, "pg instance still listed after drop")
}

func Test_PgaasLifecycle(t *testing.T) {
	suite.Run(t, new(PgaasLifecycleSuite))
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
// its per-instance data volume (pgaas.pgVolumeName is "<name>-data"),
// ignoring "no such container/volume".
func removePgaasInstance(dockerClient client.APIClient) {
	ctx := context.Background()
	removeOpts := container.RemoveOptions{Force: true}

	_ = dockerClient.ContainerRemove(ctx, pgaasLifecycleInstanceName, removeOpts)
	_ = dockerClient.VolumeRemove(ctx, pgaasLifecycleInstanceName+"-data", true)
}
