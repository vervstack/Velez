package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/storage/environments"
)

// ServiceScopingSuite covers the PROD-only proto gap fixed in Stage 2 of
// docs/container_runtimes: StopService/RestartService/RemoveService/
// GetServiceMetrics now carry an `environment` field and route through
// ContainerRuntime.{Stop,Restart,Stats}, so a same-named smerd in a
// non-default environment is no longer invisible to (or, worse, silently
// operated on in place of the PROD one by) these four RPCs.
//
// Same real-Docker, single-node (local_storage) setup as EnvironmentsSuite -
// see that suite's doc comment for why WithMatreshka() is never involved
// here.
type ServiceScopingSuite struct {
	suite.Suite
}

const (
	// serviceScopingSuffix is deliberately non-empty for the same reason
	// EnvironmentsSuite's e2eDefaultSuffix is: an empty ContainerSuffix makes
	// the suffixed and bare forms of a container name identical, which would
	// hide a real suffix-handling bug behind a passing test.
	serviceScopingSuffix = "e2esvcprod"
	serviceScopingStage  = "E2ESVCSTAGE"

	// Container names double as Docker hostnames (64-char cap), so these stay
	// short and suite-unique - same constraint EnvironmentsSuite documents.
	serviceScopingStageName = "e2e_svc_stage"
	serviceScopingSameName  = "e2e_svc_same"
)

func (s *ServiceScopingSuite) containerRunning(env *TestEnvironment, id string) bool {
	t := s.T()
	t.Helper()

	inspected, err := env.Custom.NodeClients.Docker().Client().ContainerInspect(t.Context(), id)
	require.NoError(t, err)

	return inspected.State.Running
}

// StopService scoped to a non-default environment must stop that
// environment's own container.
func (s *ServiceScopingSuite) Test_StopService_ScopedToEnvironment_StopsOwnContainer() {
	t := s.T()

	env := NewEnvironment(t,
		WithContainerSuffix(serviceScopingSuffix),
		WithEnvironments([]string{serviceScopingStage}))

	createReq := &velez_api.CreateSmerd_Request{
		Name:         serviceScopingStageName,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
		Environment:  serviceScopingStage,
	}

	created := env.CreateSmerd(t, createReq)
	require.Equal(t, velez_api.Smerd_running, created.GetStatus())

	stopReq := &velez_api.StopService_Request{
		Name:        serviceScopingStageName,
		Environment: serviceScopingStage,
	}

	_, err := env.ServiceApiClient().StopService(t.Context(), stopReq)
	require.NoError(t, err)

	require.False(t, s.containerRunning(env, created.GetUuid()),
		"StopService scoped to STAGE must stop the STAGE container")
}

// RestartService scoped to a non-default environment must restart that
// environment's own container (still running afterwards - ContainerRestart
// blocks until the container is back up).
func (s *ServiceScopingSuite) Test_RestartService_ScopedToEnvironment_RestartsOwnContainer() {
	t := s.T()

	env := NewEnvironment(t,
		WithContainerSuffix(serviceScopingSuffix),
		WithEnvironments([]string{serviceScopingStage}))

	createReq := &velez_api.CreateSmerd_Request{
		Name:         serviceScopingStageName,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
		Environment:  serviceScopingStage,
	}

	created := env.CreateSmerd(t, createReq)
	require.Equal(t, velez_api.Smerd_running, created.GetStatus())

	restartReq := &velez_api.RestartService_Request{
		Name:        serviceScopingStageName,
		Environment: serviceScopingStage,
	}

	_, err := env.ServiceApiClient().RestartService(t.Context(), restartReq)
	require.NoError(t, err)

	require.True(t, s.containerRunning(env, created.GetUuid()),
		"RestartService scoped to STAGE must leave the STAGE container running")
}

// The core Stage 2 bug fix, mirroring
// EnvironmentsSuite.Test_DropSmerd_ByUuid_CrossEnvironmentCollision's exact
// shape (a single create in STAGE, then a PROD-scoped operation on the same
// name/uuid) rather than creating two same-named smerds in one test: the
// jobs engine dedups create_smerd tasks without an environment component
// today (see EnvironmentsSuite's t.Skip'd
// Test_SameNameInTwoEnvironments_AreDistinctContainers), so two back-to-back
// same-named creates in different environments is a separate, out-of-scope
// bug this test must not depend on.
//
// A STAGE-only smerd must be completely invisible to (and untouched by) a
// PROD-scoped StopService for the same name: ListSmerds scoped to PROD finds
// nothing, StopService has nothing to stop and returns success, and the
// STAGE container is left running throughout.
func (s *ServiceScopingSuite) Test_StopService_SameNameOtherEnvironment_DoesNotTouchIt() {
	t := s.T()

	env := NewEnvironment(t,
		WithContainerSuffix(serviceScopingSuffix),
		WithEnvironments([]string{serviceScopingStage}))

	stageReq := &velez_api.CreateSmerd_Request{
		Name:         serviceScopingSameName,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
		Environment:  serviceScopingStage,
	}
	stageSmerd := env.CreateSmerd(t, stageReq)
	require.Equal(t, velez_api.Smerd_running, stageSmerd.GetStatus())

	stopReq := &velez_api.StopService_Request{
		Name:        serviceScopingSameName,
		Environment: environments.DefaultEnvironmentName,
	}

	_, err := env.ServiceApiClient().StopService(t.Context(), stopReq)
	require.NoError(t, err, "nothing to stop in PROD must still be a successful no-op")

	require.True(t, s.containerRunning(env, stageSmerd.GetUuid()),
		"a PROD-scoped StopService must NOT touch the same-named STAGE container")
}

func Test_ServiceScoping(t *testing.T) {
	suite.Run(t, new(ServiceScopingSuite))
}
