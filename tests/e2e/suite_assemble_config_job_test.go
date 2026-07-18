package e2e

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.redsock.ru/toolbox"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
	"go.vervstack.ru/Velez/tests/config_mocks"
)

// AssembleConfigJobSuite exercises the "assemble_config" job (internal/jobs/assemble_config.go)
// directly through the real JobsEngine/TaskWorker, in parallel to AssembleConfigSuite which
// covers the same scenario through the still-live pipeline. The job isn't cut over to any RPC
// yet, so it's driven here the way a future gRPC handler would: Enqueue then Watch.
type AssembleConfigJobSuite struct {
	suite.Suite

	ctx context.Context
}

func (s *AssembleConfigJobSuite) SetupSuite() {
	s.ctx = context.Background()
}

func (s *AssembleConfigJobSuite) Test_AssembleHelloWorld() {
	t := s.T()

	serviceName := GetServiceName(t)
	env := NewEnvironment(t)

	initialContext := &velez_api.AssembleConfigTaskPayload{
		ServiceName: serviceName,
		ImageName:   HelloWorldAppImage,
	}

	_, err := env.Custom.JobsEngine.Enqueue(s.ctx, serviceName, jobs.AssembleConfigAction, initialContext)
	require.NoError(t, err)

	var finalTask tasks_queries.VelezTask
	for task := range env.Custom.JobsEngine.Watch(s.ctx, serviceName, jobs.AssembleConfigAction) {
		finalTask = task
	}

	require.Equal(t, tasks_queries.VelezTaskStatusDONE, finalTask.Status, "task error: %s", finalTask.Error.String)

	var payload velez_api.AssembleConfigTaskPayload
	require.True(t, finalTask.Context.Valid)
	err = json.Unmarshal(finalTask.Context.RawMessage, &payload)
	require.NoError(t, err)

	require.Equal(t, serviceName, payload.GetConfigName())
	require.Equal(t, "verv", payload.GetConfType())
	require.NotEmpty(t, payload.GetContainerId())
	require.YAMLEq(t, string(config_mocks.HelloWorld), string(payload.GetContentRaw()))

	scratchName := serviceName + "_config_scanning"
	listReq := &velez_api.ListSmerds_Request{Name: toolbox.ToPtr(scratchName)}
	cont, err := env.Custom.NodeClients.Docker().ListContainers(s.ctx, listReq)
	require.NoError(t, err)
	require.Empty(t, cont, "scratch container should have been dropped by the job")
}

func Test_AssembleConfigJob(t *testing.T) {
	suite.Run(t, new(AssembleConfigJobSuite))
}
