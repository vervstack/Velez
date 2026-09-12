//go:build e2e_full

package e2e

import (
	"strings"
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

	// svcLifecycleRunningTimeout bounds the wait for the deploy watcher (5s
	// tick) to enqueue a create_smerd/upgrade_smerd task and for that task to
	// pull the image and bring the container up.
	svcLifecycleRunningTimeout = 90 * time.Second

	// svcLifecycleUpgradeImage is a different, real hello_world tag so the
	// upgrade leg's effect is observable on the running container's image.
	svcLifecycleUpgradeImage = helloWorldImageV0015
)

// ServiceLifecycleSuite drives the full service/deployment lifecycle against
// real infrastructure once statefull_pg is enabled:
//
//	enableStatefullPgUnderDind (cluster Postgres via the ClusterPgDsn host
//	  seam, shared with EnableStatefullSuite)
//	-> CreateService              (validate_name + upsert_service jobs)
//	-> CreateDeploy(New)          (SCHEDULED_DEPLOYMENT row + specification)
//	-> deploy watcher drives it   (create_smerd task) -> RUNNING container
//	-> CreateDeploy(Upgrade)      (SCHEDULED_UPGRADE row against a new image)
//	-> deploy watcher drives it   (upgrade_smerd task) -> RUNNING new image
//
// The deploy-watcher-driven half is only observable in one process because the
// storage-binding fix ([Jobs] Fix: resolve storage backend live ...): the
// watcher and the create_service handler now hold the swappable cluster state
// manager and re-resolve their storage per call, so an "enable statefull, then
// deploy, same process" flow sees the cluster-Postgres deployment. Before that
// fix they captured the pre-swap local_storage backend at startup and the
// watcher never saw the row.
//
// Not t.Parallel(): enableStatefullPgUnderDind t.Chdir's to the repo root
// (goose migrations resolve "./migrations" relative to cwd).
type ServiceLifecycleSuite struct {
	suite.Suite

	plane Plane
}

func (s *ServiceLifecycleSuite) Test_ServiceDeploymentLifecycle() {
	t := s.T()
	ctx := t.Context()

	env, _ := enableStatefullPgUnderDind(t, s.plane, svcLifecycleSuffix)

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
	deployed := s.awaitDeployment(env, func(d *velez_api.DeploymentInfo) bool {
		return d.GetSpecId() != 0
	}, svcLifecycleWaitTimeout, "the new deployment must be listed for the service via the ServiceName filter")

	require.NotZero(t, deployed.GetId())
	require.NotZero(t, deployed.GetSpecId(), "the deployment must reference its persisted specification")
	require.Contains(t,
		[]velez_api.DeploymentStatus{
			velez_api.DeploymentStatus_SCHEDULED_DEPLOYMENT,
			velez_api.DeploymentStatus_RUNNING,
		},
		deployed.GetStatus(),
		"a freshly created deployment is scheduled, or already being driven to RUNNING by the watcher")

	newSpecId := deployed.GetSpecId()

	// The deploy watcher (create_smerd task) must drive the deployment to
	// RUNNING and leave a real, running container behind.
	s.awaitDeploymentStatus(env, deployed.GetId(), velez_api.DeploymentStatus_RUNNING,
		"the deploy watcher must drive the new deployment to RUNNING")

	// The deployed container carries the bare service name: svcLifecycleSuffix
	// isolates the cluster-state pg sidecar (state.PgName), not the smerd. A
	// cluster-mode deploy through the default environment gets no per-node
	// ContainerSuffix - the post-enable_statefull Postgres environments storage
	// seeds the default environment with an empty suffix.
	containerName := svcLifecycleServiceName

	inspected, inspectErr := env.Custom.NodeClients.Docker().Client().ContainerInspect(ctx, containerName)
	require.NoError(t, inspectErr, "the deployed container %q must exist", containerName)
	require.NotNil(t, inspected.State)
	require.True(t, inspected.State.Running, "the deployed container must be running")

	// Upgrade leg: schedule an upgrade onto a different image and let the
	// watcher (upgrade_smerd task) drive it.
	upgradeSpec := &velez_api.CreateDeploy_Request_Upgrade{
		DeploymentId: deployed.GetId(),
		Image:        toolbox.ToPtr(svcLifecycleUpgradeImage),
	}
	upgradeReq := &velez_api.CreateDeploy_Request{
		ServiceName:   svcLifecycleServiceName,
		Environment:   environments.DefaultEnvironmentName,
		Specification: &velez_api.CreateDeploy_Request_Upgrade_{Upgrade: upgradeSpec},
	}

	_, err = env.Custom.ServiceApiImpl.CreateDeploy(ctx, upgradeReq)
	require.NoError(t, err, "CreateDeploy(Upgrade) must schedule an upgrade deployment")

	upgradeDep := s.awaitDeployment(env, func(d *velez_api.DeploymentInfo) bool {
		return d.GetSpecId() != 0 && d.GetSpecId() != newSpecId
	}, svcLifecycleWaitTimeout, "the upgrade must add a deployment carrying a fresh specification")

	require.Contains(t,
		[]velez_api.DeploymentStatus{
			velez_api.DeploymentStatus_SCHEDULED_UPGRADE,
			velez_api.DeploymentStatus_RUNNING,
		},
		upgradeDep.GetStatus(),
		"the upgrade deployment is scheduled, or already being driven to RUNNING by the watcher")

	s.awaitDeploymentStatus(env, upgradeDep.GetId(), velez_api.DeploymentStatus_RUNNING,
		"the deploy watcher must drive the upgrade deployment to RUNNING")

	require.Eventually(t, func() bool {
		afterUpgrade, err := env.Custom.NodeClients.Docker().Client().ContainerInspect(ctx, containerName)
		if err != nil {
			return false
		}

		if afterUpgrade.State == nil || !afterUpgrade.State.Running {
			return false
		}

		return strings.Contains(afterUpgrade.Config.Image, "v0.0.15")
	}, svcLifecycleRunningTimeout, svcLifecyclePollEvery,
		"after the upgrade the container must still be running, now on the upgraded image")
}

// awaitDeployment polls the real ListDeployments RPC (ServiceName filter) until
// one listed deployment satisfies match, and returns it.
func (s *ServiceLifecycleSuite) awaitDeployment(
	env *TestEnvironment,
	match func(d *velez_api.DeploymentInfo) bool,
	timeout time.Duration,
	msg string,
) *velez_api.DeploymentInfo {
	t := s.T()
	ctx := t.Context()

	var found *velez_api.DeploymentInfo

	require.Eventually(t, func() bool {
		req := &velez_api.ListDeployments_Request{ServiceName: toolbox.ToPtr(svcLifecycleServiceName)}

		resp, listErr := env.Custom.ServiceApiImpl.ListDeployments(ctx, req)
		if listErr != nil {
			return false
		}

		for _, d := range resp.GetDeployments() {
			if match(d) {
				found = d

				return true
			}
		}

		return false
	}, timeout, svcLifecyclePollEvery, msg)

	return found
}

// awaitDeploymentStatus polls the real ListDeployments RPC until the deployment
// with id reaches want, failing if it lands on FAILED instead.
func (s *ServiceLifecycleSuite) awaitDeploymentStatus(
	env *TestEnvironment,
	id uint64,
	want velez_api.DeploymentStatus,
	msg string,
) {
	t := s.T()
	ctx := t.Context()

	var last velez_api.DeploymentStatus

	require.Eventually(t, func() bool {
		req := &velez_api.ListDeployments_Request{ServiceName: toolbox.ToPtr(svcLifecycleServiceName)}

		resp, listErr := env.Custom.ServiceApiImpl.ListDeployments(ctx, req)
		if listErr != nil {
			return false
		}

		for _, d := range resp.GetDeployments() {
			if d.GetId() != id {
				continue
			}

			last = d.GetStatus()

			return last == want || last == velez_api.DeploymentStatus_FAILED
		}

		return false
	}, svcLifecycleRunningTimeout, svcLifecyclePollEvery, msg)

	require.Equal(t, want, last, msg)
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
}

func Test_ServiceLifecycle(t *testing.T) {
	RunPlaneSuite(t, ClusterPlanes, func(plane Plane) suite.TestingSuite {
		return &ServiceLifecycleSuite{plane: plane}
	})
}
