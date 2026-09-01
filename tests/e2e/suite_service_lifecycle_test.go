package e2e

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.redsock.ru/toolbox"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/storage/environments"
)

const (
	// svcLifecycleSuffix is this suite's own ContainerSuffix - distinct from
	// EnableStatefullSuite's and from production's empty default - so the
	// cluster-pg sidecar (state.PgName(svcLifecycleSuffix)) can never collide
	// with, or be force-removed alongside, another suite's or a developer's
	// containers.
	svcLifecycleSuffix = "e2e-svc-lifecycle"

	// svcLifecycleServiceName is the service name. >= 4 chars and [A-Za-z0-9_]
	// only, per jobs.minServiceNameLen / allowedServiceNameSymbols. Kept short:
	// container names double as Docker hostnames (64-char cap).
	svcLifecycleServiceName = "e2e_svc_lifecycle"

	svcLifecycleWaitTimeout = 30 * time.Second
	svcLifecyclePollEvery   = 2 * time.Second
)

// ServiceLifecycleSuite drives the service/deployment RPCs against real
// infrastructure once statefull_pg is enabled: it stands up the cluster
// Postgres through the Phase C ClusterPgDsn host seam (enableStatefullPgUnderDind,
// shared with EnableStatefullSuite), then exercises CreateService and
// CreateDeploy(New) through the real ServiceApi handlers and asserts the
// deployment + specification land in the cluster Postgres, reachable back
// through ListDeployments' ServiceName join.
//
// SCOPE NOTE / TODO(#125): the deploy-watcher-driven half of the lifecycle -
// SCHEDULED_DEPLOYMENT -> RUNNING container, then CreateDeploy(Upgrade) ->
// transition - is NOT covered here and cannot be from a single in-process
// test. internal/app/custom.go constructs both the deploy watcher
// (workers.NewDeployWatcher, via clusterClients.StateManager().Deployments())
// and the create_service handler (jobs.NewCreateServiceHandler, via
// StateManager().Services()) by resolving the storage backend ONCE at
// startup - i.e. the pre-swap local_storage backend. enable_statefull swaps
// the backend under the container atomic pointer, but those two already hold
// the old concrete storage, so in a "enable statefull, then deploy, same
// process" flow the watcher never observes the cluster-Postgres deployment.
// In production the node restarts with cluster storage already active before
// the watcher starts, so this only bites the single-process test. Making
// those two resolve storage live (the fix VervService already carries) is a
// product change outside the Phase C cluster-pg host seam.
//
// Not t.Parallel(): enableStatefullPgUnderDind t.Chdir's to the repo root
// (goose migrations resolve "./migrations" relative to cwd).
type ServiceLifecycleSuite struct {
	suite.Suite
}

func (s *ServiceLifecycleSuite) Test_ServiceDeploymentLifecycle() {
	t := s.T()
	ctx := t.Context()

	env, _ := enableStatefullPgUnderDind(t, svcLifecycleSuffix)

	s.createService(env)

	newSpec := &velez_api.CreateSmerd_Request{
		Name:         svcLifecycleServiceName,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
		Labels:       map[string]string{testCaseNameLabel: t.Name()},
	}
	deployReq := &velez_api.CreateDeploy_Request{
		ServiceName:   svcLifecycleServiceName,
		Environment:   environments.DefaultEnvironmentName,
		Specification: &velez_api.CreateDeploy_Request_New{New: newSpec},
	}

	_, err := env.Custom.ServiceApiImpl.CreateDeploy(ctx, deployReq)
	require.NoError(t, err, "CreateDeploy(New) must persist a scheduled deployment against the cluster postgres")

	// The deployment + its specification must be readable back through the
	// real ListDeployments RPC, filtered by ServiceName - the clause that
	// exercises the deployment_specifications.service_id join against real
	// Postgres.
	var deployed *velez_api.DeploymentInfo

	require.Eventually(t, func() bool {
		req := &velez_api.ListDeployments_Request{ServiceName: toolbox.ToPtr(svcLifecycleServiceName)}

		resp, listErr := env.Custom.ServiceApiImpl.ListDeployments(ctx, req)
		if listErr != nil || len(resp.GetDeployments()) == 0 {
			return false
		}

		deployed = resp.GetDeployments()[0]

		return true
	}, svcLifecycleWaitTimeout, svcLifecyclePollEvery,
		"the new deployment must be listed for the service via the ServiceName filter")

	require.NotZero(t, deployed.GetId())
	require.NotZero(t, deployed.GetSpecId(), "the deployment must reference its persisted specification")
	require.Equal(t, velez_api.DeploymentStatus_SCHEDULED_DEPLOYMENT, deployed.GetStatus(),
		"a freshly created deployment is scheduled; see the suite SCOPE NOTE on why it is not driven to RUNNING here")

	t.Log("TODO(#125): SCHEDULED_DEPLOYMENT -> RUNNING container and CreateDeploy(Upgrade) are not asserted - " +
		"the deploy watcher and create_service handler bind their storage backend at startup, before the " +
		"enable_statefull swap, so an in-process enable-then-deploy flow cannot observe the cluster-postgres " +
		"deployment. See the suite doc comment.")
}

func (s *ServiceLifecycleSuite) createService(env *TestEnvironment) {
	t := s.T()
	ctx := t.Context()

	req := &velez_api.CreateService_Request{
		Name:        svcLifecycleServiceName,
		Environment: environments.DefaultEnvironmentName,
	}

	_, err := env.Custom.ServiceApiImpl.CreateService(ctx, req)
	require.NoError(t, err, "the CreateService RPC (validate_name + upsert_service jobs) must not error")

	// The CreateService RPC above writes through the create_service handler,
	// which internal/app/custom.go bound to the pre-swap local_storage backend
	// at startup (see the suite SCOPE NOTE). Seed the service directly through
	// the live storage container the deploy path reads so CreateDeploy can
	// resolve it in the cluster postgres.
	err = env.Custom.Services.StorageContainer().Services().UpsertService(ctx, svcLifecycleServiceName)
	require.NoError(t, err, "seeding the service into the post-swap cluster storage must succeed")
}

func Test_ServiceLifecycle(t *testing.T) {
	suite.Run(t, new(ServiceLifecycleSuite))
}
