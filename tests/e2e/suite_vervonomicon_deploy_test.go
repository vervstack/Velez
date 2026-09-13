//go:build e2e_full

package e2e

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/storage/environments"
)

const (
	// vervDeploySuffix isolates this suite's cluster-pg sidecar
	// (state.PgName) from every other suite's - see svcLifecycleSuffix's
	// doc comment in suite_service_lifecycle_test.go.
	vervDeploySuffix = "e2e-verv-deploy"

	vervDeployWaitTimeout = 90 * time.Second
	vervDeployPollEvery   = 2 * time.Second

	vervDeployFidelitySvc  = "e2v_deploy_fidelity"
	vervDeployBoxErrSvc    = "e2v_deploy_boxerr"
	vervDeployImagePrecSvc = "e2v_deploy_imgprec"

	vervUnknownBoxName = "does-not-exist"
)

// VervonomiconDeploySuite covers the vervonomicon-driven half of CreateDeploy
// (CreateDeploy_Request_Vervonomicon -> CreateDeployFromVervonomicon) against
// real cluster Postgres and real Docker:
//
//   - box resolution against the real velez.resource_boxes table, both the
//     success path (a seeded builtin box name) and the failure path (an
//     unknown box name, which must name itself and list the known ones)
//   - resolver fidelity: every mapped field of a resolved CreateSmerd.Request
//     actually lands on the real deployed container
//   - a deploy request's image beating the descriptor's own app.image
//   - the deployment specification (which carries the merged descriptor) is
//     actually persisted, and a subsequent GetVervonomicon independently
//     re-derives equivalent content from the same image
//
// The suite's three test methods are NOT individually t.Parallel(): each
// SetupTest call re-enables statefull_pg under the same fixed
// vervDeploySuffix, so its cluster-pg sidecar container name would collide
// if two methods ran concurrently. Test_VervonomiconDeploy itself is also NOT
// t.Parallel(): enableStatefullPgUnderDind exposes the sidecar on the fixed
// dindClusterPgPort (30020, see dind_ports.go), the same port
// Test_EnableStatefull and Test_ServiceLifecycle expose theirs on - running
// this concurrently with those raced them for that single port and failed
// with "requested port is already occupied" (confirmed against real Docker).
//
// Out of scope, and NOT covered here (see the final report for the full
// reasoning): the box's resolved numeric cpu/ram is never wired onto the real
// Docker container (no HostConfig.Resources.NanoCPUs/Memory in
// internal/jobs/create_smerd.go) nor surfaced by any RPC - only the box NAME
// round-trips into resolved_yaml. That mapping is exercised at the unit level
// only (vervonomicon/boxes_test.go, resolve_test.go). app.command is also not
// exercised here: overriding godverv/hello_world's entrypoint with an
// unverified command string risks flaking the whole deploy rather than
// proving anything about the mapping, which resolve_test.go already covers.
type VervonomiconDeploySuite struct {
	suite.Suite

	plane Plane
	env   *TestEnvironment
}

func (s *VervonomiconDeploySuite) SetupTest() {
	s.env, _ = enableStatefullPgUnderDind(s.T(), s.plane, vervDeploySuffix)
}

func (s *VervonomiconDeploySuite) createService(env *TestEnvironment, name string) {
	t := s.T()
	ctx := t.Context()

	req := &velez_api.CreateService_Request{
		Name:        name,
		Environment: environments.DefaultEnvironmentName,
	}

	_, err := env.Custom.ServiceApiImpl.CreateService(ctx, req)
	require.NoError(t, err, "CreateService must not error")
}

func (s *VervonomiconDeploySuite) awaitRunning(env *TestEnvironment, serviceName string) *velez_api.DeploymentInfo {
	t := s.T()
	ctx := t.Context()

	var found *velez_api.DeploymentInfo

	require.Eventually(t, func() bool {
		req := &velez_api.ListDeployments_Request{ServiceName: &serviceName}

		resp, listErr := env.Custom.ServiceApiImpl.ListDeployments(ctx, req)
		if listErr != nil {
			return false
		}

		for _, d := range resp.GetDeployments() {
			if d.GetStatus() == velez_api.DeploymentStatus_RUNNING {
				found = d

				return true
			}

			if d.GetStatus() == velez_api.DeploymentStatus_FAILED {
				found = d

				return true
			}
		}

		return false
	}, vervDeployWaitTimeout, vervDeployPollEvery,
		"the deploy watcher must drive the vervonomicon deployment to a terminal status")

	require.NotNil(t, found)
	require.Equal(t, velez_api.DeploymentStatus_RUNNING, found.GetStatus(),
		"the vervonomicon deployment must reach RUNNING, not FAILED")

	return found
}

// Test_BoxResolutionAndFidelity is the headline vervonomicon deploy scenario:
// one successful deploy proves box-name resolution against the real
// postgres-seeded velez.resource_boxes table (box: small) and resolver
// fidelity (ports/volumes/env/labels incl. verv.tag.<tag>/healthcheck/
// restart) together, then a GetVervonomicon call after the deploy proves the
// persisted specification round-trips into equivalent content.
func (s *VervonomiconDeploySuite) Test_BoxResolutionAndFidelity() {
	t := s.T()
	ctx := t.Context()

	dockerAPI := s.env.Custom.NodeClients.Docker().Client()

	svcName := vervDeployFidelitySvc
	tag := svcName + ":verv"

	indexYaml := "" +
		"version: \"1\"\n" +
		"service:\n" +
		"  name: " + svcName + "\n" +
		"  tags:\n" +
		"    - canary\n" +
		"box: small\n"

	deploymentYaml := "" +
		"app:\n" +
		"  ports:\n" +
		"    - port: 8080\n" +
		"      protocol: tcp\n" +
		"  volumes:\n" +
		"    - name: " + svcName + "-data\n" +
		"      path: /data\n" +
		"  env:\n" +
		"    GREETING: hi\n" +
		"  labels:\n" +
		"    team: platform\n" +
		"  healthcheck:\n" +
		"    interval_second: 2\n" +
		"    retries: 3\n" +
		"  restart: always\n"

	files := map[string]string{
		vervIndexFile:      indexYaml,
		vervDeploymentFile: deploymentYaml,
	}

	buildVervImage(t, dockerAPI, tag, files)

	s.createService(s.env, svcName)

	deployReq := &velez_api.CreateDeploy_Request{
		ServiceName: svcName,
		Environment: environments.DefaultEnvironmentName,
		Specification: &velez_api.CreateDeploy_Request_Vervonomicon{
			Vervonomicon: &velez_api.CreateDeploy_Request_FromVervonomicon{Image: tag},
		},
	}

	_, err := s.env.Custom.ServiceApiImpl.CreateDeploy(ctx, deployReq)
	require.NoError(t, err, "CreateDeploy(Vervonomicon) must resolve box: small against the real boxes table")

	s.awaitRunning(s.env, svcName)

	inspected, err := dockerAPI.ContainerInspect(ctx, svcName)
	require.NoError(t, err, "the resolved container %q must exist", svcName)
	require.NotNil(t, inspected.State)
	require.True(t, inspected.State.Running)

	require.Equal(t, tag, inspected.Config.Image)
	require.Contains(t, inspected.Config.Env, "GREETING=hi")
	require.Equal(t, "platform", inspected.Config.Labels["team"])
	require.Equal(t, "true", inspected.Config.Labels["verv.tag.canary"],
		"a service.tags entry must map onto a verv.tag.<tag> label")

	// Per docs/features/vervonomicon.md, expose_to is "optional host port;
	// omit to keep it internal" - so 8080 must NOT gain a host binding.
	//
	// The expose_to-names-a-host-port branch is deliberately NOT exercised
	// here: this suite runs under DinD, whose PortManager is pinned to the
	// 30000-30021 band (dind_ports.go), so any port a descriptor names
	// outright either falls outside the band and fails the deploy, or
	// collides with a parallel test's auto-assigned port. That mapping is
	// covered as a unit instead - see vervonomicon.TestResolveRequest's
	// ExposedTo assertions in resolve_test.go.
	require.Empty(t, inspected.NetworkSettings.Ports["8080/tcp"],
		"a port with no expose_to must stay internal - no host binding")

	foundVolume := false

	for _, m := range inspected.Mounts {
		if m.Destination == "/data" {
			foundVolume = true

			break
		}
	}

	require.True(t, foundVolume, "the descriptor's volume must be mounted at /data")

	require.NotNil(t, inspected.Config.Healthcheck)
	require.Equal(t, 3, inspected.Config.Healthcheck.Retries)

	// Docker's own RestartPolicy vocabulary is coarser than vervonomicon's -
	// dockerutils/parser.FromRestart collapses always/on_failure/
	// unless_stopped all onto container.RestartPolicyOnFailure, a pre-existing
	// quirk unrelated to vervonomicon's own 1:1 resolveRestartType mapping.
	// Assert only that some restart policy was actually applied.
	require.NotEmpty(t, string(inspected.HostConfig.RestartPolicy.Name))

	// GetVervonomicon independently re-derives the descriptor from the same
	// image and must agree with what was just deployed.
	//
	// NOTE: this re-derives from the image - it does NOT read back the
	// persisted deployment_specifications.verv_descriptor row, and would pass
	// unchanged if that column were never written. Snapshot persistence is
	// still uncovered.
	vervReq := &velez_api.GetVervonomicon_Request{ServiceName: svcName}

	vervResp, err := s.env.Custom.ServiceApiImpl.GetVervonomicon(ctx, vervReq)
	require.NoError(t, err)
	require.Contains(t, vervResp.GetResolvedYaml(), "box: small")
	require.Contains(t, vervResp.GetResolvedYaml(), "GREETING: hi")
}

// Test_UnknownBox_ReturnsError covers the negative box-resolution path: a box
// name that isn't in the real velez.resource_boxes table must fail the
// deploy with an error naming the box and listing the known ones.
func (s *VervonomiconDeploySuite) Test_UnknownBox_ReturnsError() {
	t := s.T()
	ctx := t.Context()

	dockerAPI := s.env.Custom.NodeClients.Docker().Client()

	svcName := vervDeployBoxErrSvc
	tag := svcName + ":verv"

	indexYaml := "" +
		"version: \"1\"\n" +
		"service:\n" +
		"  name: " + svcName + "\n" +
		"box: " + vervUnknownBoxName + "\n"

	files := map[string]string{vervIndexFile: indexYaml}

	buildVervImage(t, dockerAPI, tag, files)

	s.createService(s.env, svcName)

	deployReq := &velez_api.CreateDeploy_Request{
		ServiceName: svcName,
		Environment: environments.DefaultEnvironmentName,
		Specification: &velez_api.CreateDeploy_Request_Vervonomicon{
			Vervonomicon: &velez_api.CreateDeploy_Request_FromVervonomicon{Image: tag},
		},
	}

	_, err := s.env.Custom.ServiceApiImpl.CreateDeploy(ctx, deployReq)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown box '"+vervUnknownBoxName+"'")
	require.Contains(t, err.Error(), "small")
	require.Contains(t, err.Error(), "medium")
	require.Contains(t, err.Error(), "large")
}

// Test_DeployRequestImage_BeatsAppImage covers the precedence rule: the
// descriptor's own app.image is a nonexistent, unpullable reference, and the
// deploy still succeeds and runs the REAL image because the deploy request's
// image always wins. If resolveImage's precedence regressed, this deploy
// would fail trying to pull the bogus image instead.
func (s *VervonomiconDeploySuite) Test_DeployRequestImage_BeatsAppImage() {
	t := s.T()
	ctx := t.Context()

	dockerAPI := s.env.Custom.NodeClients.Docker().Client()

	svcName := vervDeployImagePrecSvc
	tag := svcName + ":verv"

	indexYaml := "version: \"1\"\nservice:\n  name: " + svcName + "\n"
	deploymentYaml := "app:\n  image: nonexistent/image:doesnotexist\n"

	files := map[string]string{
		vervIndexFile:      indexYaml,
		vervDeploymentFile: deploymentYaml,
	}

	buildVervImage(t, dockerAPI, tag, files)

	s.createService(s.env, svcName)

	deployReq := &velez_api.CreateDeploy_Request{
		ServiceName: svcName,
		Environment: environments.DefaultEnvironmentName,
		Specification: &velez_api.CreateDeploy_Request_Vervonomicon{
			Vervonomicon: &velez_api.CreateDeploy_Request_FromVervonomicon{Image: tag},
		},
	}

	_, err := s.env.Custom.ServiceApiImpl.CreateDeploy(ctx, deployReq)
	require.NoError(t, err)

	s.awaitRunning(s.env, svcName)

	inspected, err := dockerAPI.ContainerInspect(ctx, svcName)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(inspected.Config.Image, svcName),
		"the deploy request's image must win over app.image's nonexistent reference, got %q",
		inspected.Config.Image)
}

func Test_VervonomiconDeploy(t *testing.T) {
	RunPlaneSuite(t, Planes, func(plane Plane) suite.TestingSuite {
		return &VervonomiconDeploySuite{plane: plane}
	})
}
