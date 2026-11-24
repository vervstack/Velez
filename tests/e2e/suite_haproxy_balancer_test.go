//go:build e2e_full

package e2e

import (
	"context"
	"net/url"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

type HaproxyBalancerSuite struct {
	suite.Suite

	plane Plane
}

func (s *HaproxyBalancerSuite) Test_RepointsPostgresTraffic() {
	t := s.T()
	t.Parallel()

	env := s.plane.NewEnvironment(t)
	suffix := "-" + string(s.plane.Mode)
	pgA := createPgWithCleanup(t, env, "e2e-haproxy-a"+suffix)
	pgB := createPgWithCleanup(t, env, "e2e-haproxy-b"+suffix)

	runReplicationSql(t, env, pgA, "CREATE TABLE balancer_marker(v text); INSERT INTO balancer_marker VALUES ('A')")
	runReplicationSql(t, env, pgB, "CREATE TABLE balancer_marker(v text); INSERT INTO balancer_marker VALUES ('B')")

	name := "e2e-haproxy" + suffix
	port := haproxyFixturePort(t, env, pgA, pgB)

	createHaproxyAndAwait(t, env, name, pgA.GetName(), port)

	balancer := findHaproxyInstance(t, env, name)
	require.Equal(t, pgA.GetName(), balancer.GetMasterPgInstance())
	require.Equal(t, port, balancer.GetPort())

	requireHaproxyMarker(t, env, balancer.GetPort(), pgA.GetName(), "A")
	selectHaproxyMasterAndAwait(t, env, name, pgB.GetName())
	requireHaproxyMarker(t, env, balancer.GetPort(), pgB.GetName(), "B")

	updated := findHaproxyInstance(t, env, name)
	require.Equal(t, pgB.GetName(), updated.GetMasterPgInstance())

	dropHaproxyAndAwait(t, env, name)
	dropPgAndAwait(t, env, pgB.GetName())
	dropPgAndAwait(t, env, pgA.GetName())
}

//nolint:tparallel // The suite method calls t.Parallel inside RunPlaneSuite.
func Test_HaproxyBalancer(t *testing.T) {
	t.Parallel()

	RunPlaneSuite(t, Planes, func(plane Plane) suite.TestingSuite {
		return &HaproxyBalancerSuite{plane: plane}
	})
}

func haproxyFixturePort(t *testing.T, env *TestEnvironment, pgA, pgB *velez_api.PgInstance) uint32 {
	t.Helper()

	// Velez's own allocator hands out slot ports from the front of
	// AvailablePorts, so the balancer takes the last free one (the same
	// convention as reserveVelezHostPort).
	available := env.Cfg.Environment.AvailablePorts

	for i := len(available) - 1; i >= 0; i-- {
		port := uint32(available[i])
		if port != pgA.GetExposedPort() && port != pgB.GetExposedPort() {
			return port
		}
	}

	t.Fatal("no free port in the Velez fixture's port range")

	return 0
}

func createHaproxyAndAwait(t *testing.T, env *TestEnvironment, name, master string, port uint32) {
	t.Helper()

	dockerClient := env.DockerClient()
	serviceName := domain.HaproxyServiceName(name)
	volumeName := domain.HaproxyConfigVolumeName(name)
	removeOpts := container.RemoveOptions{Force: true}
	remove := func() {
		ctx := context.Background()

		_ = dockerClient.ContainerRemove(ctx, serviceName, removeOpts)
		_ = dockerClient.VolumeRemove(ctx, volumeName, true)
	}

	remove()
	t.Cleanup(remove)

	req := &velez_api.CreateHaproxyInstance_Request{
		Name:             name,
		MasterPgInstance: master,
		Port:             &port,
	}

	resp, err := env.HaproxyClient().CreateHaproxyInstance(t.Context(), req)
	require.NoError(t, err)

	task := env.AwaitTask(t, resp.GetEntityId(), resp.GetAction())
	require.Equal(t, velez_api.TaskStatus_DONE, task.GetStatus(), task.GetError())
}

func findHaproxyInstance(t *testing.T, env *TestEnvironment, name string) *velez_api.HaproxyInstance {
	t.Helper()

	req := &velez_api.ListHaproxyInstances_Request{}

	resp, err := env.HaproxyClient().ListHaproxyInstances(t.Context(), req)
	require.NoError(t, err)

	for _, instance := range resp.GetInstances() {
		if instance.GetName() == name {
			return instance
		}
	}

	t.Fatalf("expected HAProxy instance %q in ListHaproxyInstances", name)

	return nil
}

func requireHaproxyMarker(t *testing.T, env *TestEnvironment, port uint32, pgName, marker string) {
	t.Helper()

	creds := getPgCredentials(t, env, pgName)
	dsn, err := url.Parse(creds.GetDsn())
	require.NoError(t, err)

	dsn.Host = dindGatewayAddr(t, int(port))

	db := openPostgresViaDind(t, dsn.String(), env.DindDialContext)

	awaitPgReady(t, db)
	requirePgFixtureValue(t, db, "SELECT v FROM balancer_marker", marker)

	err = db.Close()
	require.NoError(t, err)
}

func selectHaproxyMasterAndAwait(t *testing.T, env *TestEnvironment, name, master string) {
	t.Helper()

	req := &velez_api.SelectHaproxyMaster_Request{Name: name, MasterPgInstance: master}

	resp, err := env.HaproxyClient().SelectHaproxyMaster(t.Context(), req)
	require.NoError(t, err)

	task := env.AwaitTask(t, resp.GetEntityId(), resp.GetAction())
	require.Equal(t, velez_api.TaskStatus_DONE, task.GetStatus(), task.GetError())
}

func dropHaproxyAndAwait(t *testing.T, env *TestEnvironment, name string) {
	t.Helper()

	req := &velez_api.DropHaproxyInstance_Request{Name: name}

	resp, err := env.HaproxyClient().DropHaproxyInstance(t.Context(), req)
	require.NoError(t, err)

	task := env.AwaitTask(t, resp.GetEntityId(), resp.GetAction())
	require.Equal(t, velez_api.TaskStatus_DONE, task.GetStatus(), task.GetError())
}
