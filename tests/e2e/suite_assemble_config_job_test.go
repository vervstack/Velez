//go:build e2e_full

package e2e

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.redsock.ru/toolbox"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/tests/config_mocks"
)

// AssembleConfigJobSuite exercises the "assemble_config" job (internal/jobs/assemble_config.go)
// through the AssembleConfig RPC and then inspects the finished task, complementing
// AssembleConfigSuite which only checks the RPC response.
type AssembleConfigJobSuite struct {
	suite.Suite

	plane Plane
	ctx   context.Context
}

func (s *AssembleConfigJobSuite) SetupSuite() {
	s.ctx = context.Background()
}

func (s *AssembleConfigJobSuite) Test_AssembleHelloWorld() {
	t := s.T()
	t.Parallel()

	serviceName := GetServiceName(t)
	env := s.plane.NewEnvironment(t)

	req := &velez_api.AssembleConfig_Request{
		ServiceName: serviceName,
		ImageName:   HelloWorldAppImage,
	}

	resp, err := env.VelezApiClient().AssembleConfig(s.ctx, req)
	require.NoError(t, err)

	finalTask := env.AwaitTask(t, serviceName, jobs.AssembleConfigAction)
	require.Equal(t, velez_api.TaskStatus_DONE, finalTask.GetStatus(), "task error: %s", finalTask.GetError())

	var payload velez_api.AssembleConfigTaskPayload

	err = json.Unmarshal([]byte(env.TaskContext(t, serviceName, jobs.AssembleConfigAction)), &payload)
	require.NoError(t, err)

	require.Equal(t, serviceName, payload.GetConfigName())
	require.Equal(t, "verv", payload.GetConfType())
	require.NotEmpty(t, payload.GetContainerId())
	require.YAMLEq(t, string(config_mocks.HelloWorld), string(payload.GetContentRaw()))
	require.YAMLEq(t, string(config_mocks.HelloWorld), string(resp.GetConfig()))

	scratchName := serviceName + "_config_scanning"
	listReq := &velez_api.ListSmerds_Request{Name: toolbox.ToPtr(scratchName)}
	cont, err := dockerutils.ListContainers(s.ctx, env.DockerClient(), listReq)
	require.NoError(t, err)
	require.Empty(t, cont, "scratch container should have been dropped by the job")
}

func Test_AssembleConfigJob(t *testing.T) {
	t.Parallel()
	RunPlaneSuite(t, Planes, func(plane Plane) suite.TestingSuite {
		return &AssembleConfigJobSuite{plane: plane}
	})
}
