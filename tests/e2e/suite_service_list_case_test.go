//go:build e2e_full

package e2e

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	// serviceListCaseSuffix is this suite's own ContainerSuffix - see every
	// other suite's identical comment on why it must be unique.
	serviceListCaseSuffix = "e2esvccase"

	// Both names are deliberately mixed-case: this is the exact VervServiceLabel
	// shape a real GitLab-runner instance carries (internal/jobs/create_runner.go
	// builds it from the caller-supplied runner name, never lowercased) - see
	// this test's doc comment. Each test method that creates a smerd gets its
	// own name - reusing one across methods races two t.Parallel() tests to
	// create the identically-named container (see serviceScopingSuffix's
	// doc comment in suite_service_scoping_test.go for the same collision).
	serviceListCaseListName = "E2E_Case_List"
	serviceListCaseGetName  = "E2E_Case_Get"
)

// ServiceListCaseSuite is the regression coverage for the ListServices /
// GetService name-corruption bug fixed in container_manager.ListSmerds:
// VervServicesService.List and .Get each enrich a domain.ServiceBaseInfo /
// domain.Service by calling ContainerManager.ListSmerds with
// &velez_api.ListSmerds_Request{Name: &svc.Name} - a pointer straight into
// their own struct field, not a copy. ListSmerds used to lowercase that
// field in place (*req.Name = strings.ToLower(...)) to normalize its own
// Docker "name" filter, which silently rewrote the caller's svc.Name too.
// A service whose real container/label name carries any uppercase letter
// (any satellite instance a caller names itself - runner, registry, pgaas)
// therefore came back from ListServices already lowercased, and the exact
// name the UI/API client then fed back into GetService no longer matched
// the real VervServiceLabel value, 404ing a service the list had just
// handed out. Reproduced live against a running node before the fix; see
// the fix commit for the manual repro.
type ServiceListCaseSuite struct {
	suite.Suite

	plane Plane
}

func newServiceListCaseCreateRequest(name string) *velez_api.CreateSmerd_Request {
	return &velez_api.CreateSmerd_Request{
		Name:         name,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
		// A plain CreateSmerd carries no VervServiceLabel of its own - every
		// satellite create flow (runner, registry, pgaas) sets it by hand to
		// the instance's own name, so this mirrors that exactly instead of
		// relying on a mechanism this test isn't targeting.
		Labels: map[string]string{labels.VervServiceLabel: name},
	}
}

// Test_ListServices_PreservesNameCase is the ListServices half: the entry
// for the just-created smerd must come back with its name byte-for-byte
// unchanged, never lowercased by the ContainerManager.ListSmerds call
// List() uses internally to enrich it.
func (s *ServiceListCaseSuite) Test_ListServices_PreservesNameCase() {
	t := s.T()
	t.Parallel()

	env := s.plane.NewEnvironment(t, WithContainerSuffix(serviceListCaseSuffix))

	created := env.CreateSmerd(t, newServiceListCaseCreateRequest(serviceListCaseListName))
	require.Equal(t, velez_api.Smerd_running, created.GetStatus())

	listReq := &velez_api.ListServices_Request{IncludeInternal: true}

	listResp, err := env.ServiceApiClient().ListServices(t.Context(), listReq)
	require.NoError(t, err)

	found := findServiceByName(listResp.GetServices(), serviceListCaseListName)
	require.NotNil(t, found, "the created service must be listed")
	require.Equal(t, serviceListCaseListName, found.GetName(),
		"ListServices must not corrupt the service's name case")
}

// Test_GetService_AfterList_FindsTheListedService is the end-to-end half
// that matches how a real caller (the UI) drives these two RPCs back to
// back: GetService, called with the exact name ListServices just handed
// back, must resolve the same service rather than 404ing against the real,
// case-sensitive VervServiceLabel on the container.
func (s *ServiceListCaseSuite) Test_GetService_AfterList_FindsTheListedService() {
	t := s.T()
	t.Parallel()

	env := s.plane.NewEnvironment(t, WithContainerSuffix(serviceListCaseSuffix))

	created := env.CreateSmerd(t, newServiceListCaseCreateRequest(serviceListCaseGetName))
	require.Equal(t, velez_api.Smerd_running, created.GetStatus())

	listReq := &velez_api.ListServices_Request{IncludeInternal: true}

	listResp, err := env.ServiceApiClient().ListServices(t.Context(), listReq)
	require.NoError(t, err)

	found := findServiceByName(listResp.GetServices(), serviceListCaseGetName)
	require.NotNil(t, found, "the created service must be listed")

	getReq := &velez_api.GetService_Request{Name: found.GetName()}

	getResp, err := env.ServiceApiClient().GetService(t.Context(), getReq)
	require.NoError(t, err, "GetService must resolve the exact name ListServices returned")
	require.Equal(t, serviceListCaseGetName, getResp.GetVervService().GetName())
}

// findServiceByName matches this suite's own service by name, case
// insensitively - the pre-fix bug returned it lowercased, and matching it
// here (rather than filtering it out) is what lets
// Test_ListServices_PreservesNameCase fail with a useful diff instead of a
// "not found" false negative.
func findServiceByName(services []*velez_api.ServiceBaseInfo, name string) *velez_api.ServiceBaseInfo {
	for _, svc := range services {
		if strings.EqualFold(svc.GetName(), name) {
			return svc
		}
	}

	return nil
}

func Test_ServiceListCase(t *testing.T) {
	t.Parallel()
	RunPlaneSuite(t, Planes, func(plane Plane) suite.TestingSuite {
		return &ServiceListCaseSuite{plane: plane}
	})
}
