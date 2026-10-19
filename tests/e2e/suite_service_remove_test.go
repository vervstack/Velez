//go:build e2e_full

package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	serviceRemoveSuffix = "e2esvcrm"

	serviceRemoveCollisionName      = "e2e_rm_base"
	serviceRemoveCollisionOtherName = "e2e_rm_base-other"
	serviceRemoveDependedName       = "e2e_rm_target"
	serviceRemoveDependentName      = "e2e_rm_caller"
)

// ServiceRemoveSuite covers RemoveService selecting containers by the exact
// VERV_SERVICE label (Docker's name filter is a substring match) and refusing
// to remove a service other services depend on.
//
// The runner -> DinD dependency is covered by
// Test_ServiceRemove_RunnerOnDind_RejectsRemoval below.
type ServiceRemoveSuite struct {
	suite.Suite

	plane Plane
}

func newServiceRemoveCreateRequest(name string, extraLabels map[string]string) *velez_api.CreateSmerd_Request {
	smerdLabels := map[string]string{labels.VervServiceLabel: name}
	for k, v := range extraLabels {
		smerdLabels[k] = v
	}

	return &velez_api.CreateSmerd_Request{
		Name:         name,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
		Labels:       smerdLabels,
	}
}

func newServiceRemoveRequest(name string) *velez_api.RemoveService_Request {
	return &velez_api.RemoveService_Request{
		Name:                 name,
		DropRunningInstances: true,
	}
}

func (s *ServiceRemoveSuite) isContainerRunning(t *testing.T, env *TestEnvironment, id string) bool {
	t.Helper()

	inspected, err := env.DockerClient().ContainerInspect(t.Context(), id)
	if err != nil {
		return false
	}

	return inspected.State.Running
}

func (s *ServiceRemoveSuite) Test_RemoveService_DoesNotDropServiceWithSimilarName() {
	t := s.T()
	t.Parallel()

	env := s.plane.NewEnvironment(t, WithContainerSuffix(serviceRemoveSuffix))

	target := env.CreateSmerd(t, newServiceRemoveCreateRequest(serviceRemoveCollisionName, nil))
	neighbour := env.CreateSmerd(t, newServiceRemoveCreateRequest(serviceRemoveCollisionOtherName, nil))

	removeReq := newServiceRemoveRequest(serviceRemoveCollisionName)

	_, err := env.ServiceApiClient().RemoveService(t.Context(), removeReq)
	require.NoError(t, err)

	require.False(t, s.isContainerRunning(t, env, target.GetUuid()),
		"the removed service's own container must be dropped")
	require.True(t, s.isContainerRunning(t, env, neighbour.GetUuid()),
		"a service whose name merely contains the removed name must be left alone")
}

func (s *ServiceRemoveSuite) Test_RemoveService_RejectedWhileOtherServiceDependsOnIt() {
	t := s.T()
	t.Parallel()

	env := s.plane.NewEnvironment(t, WithContainerSuffix(serviceRemoveSuffix))

	target := env.CreateSmerd(t, newServiceRemoveCreateRequest(serviceRemoveDependedName, nil))

	dependsOn := map[string]string{labels.DependsOnLabel: serviceRemoveDependedName}
	env.CreateSmerd(t, newServiceRemoveCreateRequest(serviceRemoveDependentName, dependsOn))

	removeReq := newServiceRemoveRequest(serviceRemoveDependedName)

	_, err := env.ServiceApiClient().RemoveService(t.Context(), removeReq)
	require.Error(t, err)
	require.Equal(t, codes.FailedPrecondition, status.Code(err))
	require.Contains(t, err.Error(), serviceRemoveDependentName)

	require.True(t, s.isContainerRunning(t, env, target.GetUuid()),
		"a service other services depend on must not be dropped")
}

func Test_ServiceRemove(t *testing.T) {
	t.Parallel()
	RunPlaneSuite(t, Planes, func(plane Plane) suite.TestingSuite {
		return &ServiceRemoveSuite{plane: plane}
	})
}

func runRemoveDindUsedByRunner(t *testing.T, env *TestEnvironment, _ Plane) {
	fixture := createRunnerOnDind(t, env)

	removeReq := newServiceRemoveRequest(fixture.dindName)

	_, err := env.ServiceApiClient().RemoveService(t.Context(), removeReq)
	require.Equal(t, codes.FailedPrecondition, status.Code(err))
	require.ErrorContains(t, err, "used by:")

	requireDindListed(t, env, fixture.dindName)
}

func Test_ServiceRemove_RunnerOnDind_RejectsRemoval(t *testing.T) {
	t.Parallel()

	runOnEveryPlane(t, runRemoveDindUsedByRunner)
}
