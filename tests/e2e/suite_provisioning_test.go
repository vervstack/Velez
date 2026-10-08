//go:build e2e_full

package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

const (
	provisioningOkInstanceName   = "e2e-prov-ok"
	provisioningFailInstanceName = "e2e-prov-fail"
	provisioningMissingEntityId  = "e2e-prov-missing"
	provisioningUnknownBox       = "e2e-no-such-box"

	provisioningPollTimeout = 3 * time.Minute
	provisioningPollEvery   = 500 * time.Millisecond
)

type ProvisioningSuite struct {
	suite.Suite

	plane Plane
}

func (s *ProvisioningSuite) Test_Provisioning_CreateTaskVisibleThenGone() {
	t := s.T()
	t.Parallel()

	env := s.plane.NewEnvironment(t)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	removeProvisioningPgInstance(dockerClient, provisioningOkInstanceName)
	t.Cleanup(func() { removeProvisioningPgInstance(dockerClient, provisioningOkInstanceName) })

	createResp := createProvisioningPg(t, env, newProvisioningPgRequest(provisioningOkInstanceName, ""))

	requireProvisioningVisibleOrListed(t, env, createResp.GetEntityId(), labels.PgaasNamePrefix+provisioningOkInstanceName)

	createTask := waitProvisioningTaskTerminal(t, env, createResp)
	require.Equal(t, tasks_queries.VelezTaskStatusDONE, createTask.Status,
		"create_pg_instance task error: %s", createTask.Error.String)

	require.Nil(t, findProvisioningTask(t, env, createResp.GetEntityId()),
		"a DONE task must vanish from provisioning")
	require.NotNil(t, findPgInstance(t, env, labels.PgaasNamePrefix+provisioningOkInstanceName),
		"the created instance must be listed")
}

func (s *ProvisioningSuite) Test_Provisioning_FailedTaskIsDismissable() {
	t := s.T()
	t.Parallel()

	env := s.plane.NewEnvironment(t)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	removeProvisioningPgInstance(dockerClient, provisioningFailInstanceName)
	t.Cleanup(func() { removeProvisioningPgInstance(dockerClient, provisioningFailInstanceName) })

	createReq := newProvisioningPgRequest(provisioningFailInstanceName, provisioningUnknownBox)
	createResp := createProvisioningPg(t, env, createReq)

	failedTask := waitProvisioningTaskTerminal(t, env, createResp)
	require.Equal(t, tasks_queries.VelezTaskStatusFAILED, failedTask.Status)

	listed := findProvisioningTask(t, env, createResp.GetEntityId())
	require.NotNil(t, listed, "a recently failed task must be listed in provisioning")
	require.Equal(t, velez_api.ProvisioningTask_FAILED, listed.GetStatus())
	require.NotEmpty(t, listed.GetError())
	require.Equal(t, jobs.CreatePgInstanceAction, listed.GetAction())

	dismissReq := newDismissTaskRequest(createResp.GetEntityId(), createResp.GetAction())

	_, err := env.Custom.TasksApiImpl.DismissTask(t.Context(), dismissReq)
	require.NoError(t, err)

	require.Nil(t, findProvisioningTask(t, env, createResp.GetEntityId()),
		"a dismissed task must be gone from provisioning")

	_, err = env.Custom.TasksApiImpl.DismissTask(t.Context(), dismissReq)
	require.Equal(t, codes.NotFound, status.Code(err))

	missingReq := newDismissTaskRequest(provisioningMissingEntityId, jobs.CreatePgInstanceAction)

	_, err = env.Custom.TasksApiImpl.DismissTask(t.Context(), missingReq)
	require.Equal(t, codes.NotFound, status.Code(err))
}

func Test_Provisioning(t *testing.T) {
	t.Parallel()
	RunPlaneSuite(t, Planes, func(plane Plane) suite.TestingSuite {
		return &ProvisioningSuite{plane: plane}
	})
}

func newProvisioningPgRequest(name, box string) *velez_api.CreatePgInstance_Request {
	req := &velez_api.CreatePgInstance_Request{Name: name}

	if box != "" {
		req.Box = &box
	}

	return req
}

func newDismissTaskRequest(entityId, action string) *velez_api.DismissTask_Request {
	return &velez_api.DismissTask_Request{
		EntityId: entityId,
		Action:   action,
	}
}

func createProvisioningPg(
	t *testing.T,
	env *TestEnvironment,
	req *velez_api.CreatePgInstance_Request,
) *velez_api.CreatePgInstance_Response {
	t.Helper()

	resp, err := env.Custom.PgaasApiImpl.CreatePgInstance(t.Context(), req)
	require.NoError(t, err)
	require.Equal(t, req.GetName(), resp.GetEntityId())
	require.Equal(t, jobs.CreatePgInstanceAction, resp.GetAction())

	return resp
}

func waitProvisioningTaskTerminal(
	t *testing.T,
	env *TestEnvironment,
	createResp *velez_api.CreatePgInstance_Response,
) tasks_queries.VelezTask {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), provisioningPollTimeout)
	defer cancel()

	var last tasks_queries.VelezTask

	for task := range env.Custom.JobsEngine.Watch(ctx, createResp.GetEntityId(), createResp.GetAction()) {
		last = task
	}

	require.NoError(t, ctx.Err(), "task did not reach a terminal state in time")

	return last
}

// requireProvisioningVisibleOrListed tolerates a task that finishes before the first poll.
func requireProvisioningVisibleOrListed(t *testing.T, env *TestEnvironment, entityId, instanceName string) {
	t.Helper()

	require.Eventually(t, func() bool {
		if findProvisioningTask(t, env, entityId) != nil {
			return true
		}

		return findPgInstance(t, env, instanceName) != nil
	}, provisioningPollTimeout, provisioningPollEvery,
		"the create task must appear in provisioning or the instance must be listed")
}

func findProvisioningTask(t *testing.T, env *TestEnvironment, entityId string) *velez_api.ProvisioningTask {
	t.Helper()

	listReq := &velez_api.ListPgInstances_Request{}

	listResp, err := env.Custom.PgaasApiImpl.ListPgInstances(t.Context(), listReq)
	require.NoError(t, err)

	for _, task := range listResp.GetProvisioning() {
		if task.GetEntityId() == entityId {
			return task
		}
	}

	return nil
}

func removeProvisioningPgInstance(dockerClient client.APIClient, name string) {
	ctx := context.Background()
	serviceName := labels.PgaasNamePrefix + name

	_ = dockerClient.ContainerRemove(ctx, serviceName, container.RemoveOptions{Force: true})
	_ = dockerClient.VolumeRemove(ctx, serviceName+"-data", true)
}
