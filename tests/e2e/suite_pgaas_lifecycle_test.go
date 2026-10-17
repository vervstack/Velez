//go:build e2e_full

package e2e

import (
	"context"
	"fmt"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/metadata"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/jobs"
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

	pgaasLifecycleRequestHost = "velez.example.test:53890"
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
	dockerClient := env.DockerClient()

	removePgaasInstance(dockerClient)
	t.Cleanup(func() { removePgaasInstance(dockerClient) })

	createReq := &velez_api.CreatePgInstance_Request{
		Name: pgaasLifecycleInstanceName,
	}

	createResp, err := env.PgaasClient().CreatePgInstance(ctx, createReq)
	require.NoError(t, err)
	require.Equal(t, pgaasLifecycleInstanceName, createResp.GetEntityId())
	require.Equal(t, jobs.CreatePgInstanceAction, createResp.GetAction())

	createTask := env.AwaitTask(t, createResp.GetEntityId(), createResp.GetAction())
	require.Equal(t, velez_api.TaskStatus_DONE, createTask.GetStatus(),
		"create_pg_instance task error: %s", createTask.GetError())

	instance := findPgInstance(t, env, pgaasLifecycleServiceName)
	require.NotNil(t, instance, "expected pg instance %q in ListPgInstances", pgaasLifecycleServiceName)
	require.Equal(t, pgaasLifecycleServiceName, instance.GetName())
	require.NotEqual(t, "velez", instance.GetName())
	require.Equal(t, "running", instance.GetStatus())
	require.NotNil(t, instance.ExposedPort, "expected a published host port in ListPgInstances")
	require.Equal(t, pgaasLifecycleInstanceName, listedServiceDisplayName(t, env, pgaasLifecycleServiceName))

	// Credentials resolve through the container-env read-through path - there
	// is no velez.secrets table in single-node mode.
	credsReq := newPgaasCredentialsRequest(pgaasLifecycleServiceName)

	credsCtx := withRequestHost(ctx, pgaasLifecycleRequestHost)

	credsResp, err := env.PgaasClient().GetPgInstanceCredentials(credsCtx, credsReq)
	require.NoError(t, err)
	require.NotEmpty(t, credsResp.GetPassword())
	require.NotEmpty(t, credsResp.GetDsn())
	require.Equal(t, instance.GetDbName(), credsResp.GetDbName())
	require.Equal(t, instance.GetUsername(), credsResp.GetUsername())

	requirePgaasConnections(t, instance, credsResp)

	dropReq := &velez_api.DropPgInstance_Request{Name: pgaasLifecycleServiceName}

	dropResp, err := env.PgaasClient().DropPgInstance(ctx, dropReq)
	require.NoError(t, err)
	require.Equal(t, pgaasLifecycleServiceName, dropResp.GetEntityId())
	require.Equal(t, jobs.DropPgInstanceAction, dropResp.GetAction())

	dropTask := env.AwaitTask(t, dropResp.GetEntityId(), dropResp.GetAction())
	require.Equal(t, velez_api.TaskStatus_DONE, dropTask.GetStatus(),
		"drop_pg_instance task error: %s", dropTask.GetError())

	dropped := findPgInstance(t, env, pgaasLifecycleServiceName)
	require.Nil(t, dropped, "pg instance still listed after drop")
}

func Test_PgaasLifecycle(t *testing.T) {
	t.Parallel()
	RunPlaneSuite(t, Planes, func(plane Plane) suite.TestingSuite {
		return &PgaasLifecycleSuite{plane: plane}
	})
}

func newPgaasCredentialsRequest(serviceName string) *velez_api.GetPgInstanceCredentials_Request {
	return &velez_api.GetPgInstanceCredentials_Request{Name: serviceName}
}

func withRequestHost(ctx context.Context, hostPort string) context.Context {
	md := metadata.Pairs("x-forwarded-host", hostPort)

	return metadata.NewOutgoingContext(ctx, md)
}

func requirePgaasConnections(
	t *testing.T, instance *velez_api.PgInstance, creds *velez_api.GetPgInstanceCredentials_Response,
) {
	t.Helper()

	inNetwork := creds.GetInNetwork()
	require.NotNil(t, inNetwork)
	require.Equal(t, creds.GetDsn(), inNetwork.GetDsn())
	require.Contains(t, inNetwork.GetDsn(), "@"+pgaasLifecycleServiceName+":5432/")
	require.NotContains(t, inNetwork.GetMaskedDsn(), creds.GetPassword())
	require.Contains(t, inNetwork.GetMaskedDsn(), "••••")

	external := creds.GetExternal()
	require.NotNil(t, external, "expected an external connection for a published port")
	require.Contains(t, external.GetDsn(), fmt.Sprintf(":%d/", instance.GetExposedPort()))
	require.Contains(t, external.GetDsn(), creds.GetPassword())
	require.NotContains(t, external.GetMaskedDsn(), creds.GetPassword())
	require.Contains(t, external.GetMaskedDsn(), "••••")
	require.Contains(t, external.GetMaskedDsn(), fmt.Sprintf(":%d/", instance.GetExposedPort()))
}

// findPgInstance returns the listed pg instance with the given name, or nil.
func findPgInstance(t *testing.T, env *TestEnvironment, name string) *velez_api.PgInstance {
	t.Helper()

	listReq := &velez_api.ListPgInstances_Request{}

	listResp, err := env.PgaasClient().ListPgInstances(t.Context(), listReq)
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
