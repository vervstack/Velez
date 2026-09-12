//go:build e2e_full

package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

const (
	vervIndexFile      = "vervonomicon.yaml"
	vervDeploymentFile = "deployment.yaml"
	vervResourcesFile  = "resources.yaml"

	vervStagingEnv = "staging"
	vervOtherEnv   = "otherenv"

	vervImgSvcName    = "e2v_verv_img"
	vervEnvSvcName    = "e2v_verv_env"
	vervResSvcName    = "e2v_verv_res"
	vervNoDescSvcName = "e2v_verv_nodesc"
	vervBadSvcName    = "e2v_verv_bad"
)

// newVervonomiconRequest is this file's one recurring request shape across
// every Test_* case: a smerd deployed off a built verv image (or
// HelloWorldAppImage), optionally into a non-default environment.
func newVervonomiconRequest(name, imageName, environment string) *velez_api.CreateSmerd_Request {
	return &velez_api.CreateSmerd_Request{
		Name:         name,
		ImageName:    imageName,
		IgnoreConfig: true,
		Environment:  environment,
	}
}

type VervonomiconSuite struct {
	suite.Suite

	plane Plane
}

// Test_ImageSourcedDescriptor covers docs/features/vervonomicon.md's
// baseline: a descriptor baked into the deployed image is read back
// byte-for-byte in Raw, resolved into resolved_yaml, and reported with
// Source == IMAGE. This also exercises ImageSource.Read (source_image.go)
// end to end - it has no dedicated test file otherwise, and services.go
// wires a real ImageSource in production.
func (s *VervonomiconSuite) Test_ImageSourcedDescriptor() {
	t := s.T()
	t.Parallel()

	env := s.plane.NewEnvironment(t)
	ctx := t.Context()
	dockerAPI := env.Custom.NodeClients.Docker().Client()

	tag := vervImgSvcName + ":verv"

	indexYaml := "version: \"1\"\nservice:\n  name: " + vervImgSvcName + "\n"
	deploymentYaml := "app:\n  env:\n    GREETING: hello\n"

	files := map[string]string{
		vervIndexFile:      indexYaml,
		vervDeploymentFile: deploymentYaml,
	}

	buildVervImage(t, dockerAPI, tag, files)

	createReq := newVervonomiconRequest(vervImgSvcName, tag, "")

	env.CreateSmerd(t, createReq)

	req := &velez_api.GetVervonomicon_Request{ServiceName: vervImgSvcName}

	resp, err := env.Custom.ServiceApiImpl.GetVervonomicon(ctx, req)
	require.NoError(t, err)
	require.Equal(t, velez_api.VervonomiconSource_VERVONOMICON_SOURCE_IMAGE, resp.GetSource())

	rawByPath := make(map[string]string, len(resp.GetRaw()))
	for _, f := range resp.GetRaw() {
		rawByPath[f.GetPath()] = string(f.GetContent())
	}

	require.Equal(t, indexYaml, rawByPath[vervIndexFile])
	require.Equal(t, deploymentYaml, rawByPath[vervDeploymentFile])

	require.Contains(t, resp.GetResolvedYaml(), "GREETING: hello")
	require.Contains(t, resp.GetResolvedYaml(), "name: "+vervImgSvcName)
}

// Test_EnvironmentOverlay covers the "Environment overlays" section: the
// SAME logical service is deployed into two Velez environments (default and
// "staging") from the same image, whose descriptor carries a base
// deployment.yaml, a "staging/" overlay, and an unrelated "otherenv/"
// directory that must never affect either resolved result.
//
// Raw is asserted to still contain every environment's files (it is the
// pre-merge view, identical regardless of the query's Environment) while the
// resolved output only ever reflects the base plus the queried environment's
// overlay - proving "ignored entirely" is a merged-output property, not a
// raw-output one.
func (s *VervonomiconSuite) Test_EnvironmentOverlay() {
	t := s.T()
	t.Parallel()

	env := s.plane.NewEnvironment(t, WithEnvironments([]string{vervStagingEnv}))
	ctx := t.Context()
	dockerAPI := env.Custom.NodeClients.Docker().Client()

	tag := vervEnvSvcName + ":verv"

	indexYaml := "version: \"1\"\nservice:\n  name: " + vervEnvSvcName + "\n"
	baseDeployment := "app:\n  env:\n    MODE: base\n    COMMON: shared\n"
	stagingDeployment := "app:\n  env:\n    MODE: staging\n"
	otherDeployment := "app:\n  env:\n    MODE: otherenv-must-never-appear\n"

	files := map[string]string{
		vervIndexFile:      indexYaml,
		vervDeploymentFile: baseDeployment,
		vervStagingEnv + "/" + vervDeploymentFile: stagingDeployment,
		vervOtherEnv + "/" + vervDeploymentFile:   otherDeployment,
	}

	buildVervImage(t, dockerAPI, tag, files)

	baseReq := newVervonomiconRequest(vervEnvSvcName, tag, "")

	env.CreateSmerd(t, baseReq)

	stagingReq := newVervonomiconRequest(vervEnvSvcName, tag, vervStagingEnv)

	env.CreateSmerd(t, stagingReq)

	defaultVervReq := &velez_api.GetVervonomicon_Request{ServiceName: vervEnvSvcName}

	defaultResp, err := env.Custom.ServiceApiImpl.GetVervonomicon(ctx, defaultVervReq)
	require.NoError(t, err)
	require.Contains(t, defaultResp.GetResolvedYaml(), "MODE: base")
	require.NotContains(t, defaultResp.GetResolvedYaml(), "MODE: staging")
	require.NotContains(t, defaultResp.GetResolvedYaml(), "otherenv-must-never-appear")

	rawPaths := make([]string, 0, len(defaultResp.GetRaw()))
	for _, f := range defaultResp.GetRaw() {
		rawPaths = append(rawPaths, f.GetPath())
	}

	require.Contains(t, rawPaths, vervStagingEnv+"/"+vervDeploymentFile,
		"raw is the pre-merge view - it must still carry every environment's files")
	require.Contains(t, rawPaths, vervOtherEnv+"/"+vervDeploymentFile)

	stagingVervReq := &velez_api.GetVervonomicon_Request{
		ServiceName: vervEnvSvcName,
		Environment: vervStagingEnv,
	}

	stagingResp, err := env.Custom.ServiceApiImpl.GetVervonomicon(ctx, stagingVervReq)
	require.NoError(t, err)
	require.Contains(t, stagingResp.GetResolvedYaml(), "MODE: staging",
		"the staging overlay must override the base value")
	require.Contains(t, stagingResp.GetResolvedYaml(), "COMMON: shared",
		"a key the overlay never mentions must survive the deep merge")
	require.NotContains(t, stagingResp.GetResolvedYaml(), "otherenv-must-never-appear",
		"a non-matching environment directory must never leak into any resolved result")
}

// Test_ResourceReconciliation covers "Resource reconciliation": a
// resources.yaml entry backed by an existing velez.service_resources binding
// (simulated here by a plain "<service>_pg" container - see
// createBoundResourceContainer) reports ALREADY_CONNECTED, and one with no
// binding and no live matreshka connection reports MUST_PROVISION.
//
// This also regression-covers the local_storage.dockerServiceResourcesStorage
// .GetResources fix in this change: it used to return the full container name
// instead of the short resource-type key, which made the bound[res.Name]
// lookup in vervonomicon.ReconcileResources never match.
//
// Needs WithMatreshka(): reconcileResources always calls
// Configurator.GetVervFromApi once a descriptor has any resources[] entry,
// even to establish "no live connection exists" for one with no binding.
func (s *VervonomiconSuite) Test_ResourceReconciliation() {
	t := s.T()
	t.Parallel()

	// WithMatreshka()'s verv://matreshka gRPC resolver still produces zero
	// addresses for in-process clients in this e2e harness (see
	// suite_verv_config_test.go's TODO), so Configurator.GetVervFromApi fails
	// here - but reconcileResources treats that as "no live connection known"
	// rather than a hard failure (see its own doc comment), so this test still
	// exercises the real decision logic end to end.
	env := s.plane.NewEnvironment(t, WithMatreshka())
	ctx := t.Context()
	dockerAPI := env.Custom.NodeClients.Docker().Client()

	tag := vervResSvcName + ":verv"

	indexYaml := "version: \"1\"\nservice:\n  name: " + vervResSvcName + "\n"
	resourcesYaml := "" +
		"- name: pg\n" +
		"  type: postgres\n" +
		"  binds_to: data_sources.postgres\n" +
		"- name: cache\n" +
		"  type: redis\n" +
		"  binds_to: data_sources.redis\n"

	files := map[string]string{
		vervIndexFile:     indexYaml,
		vervResourcesFile: resourcesYaml,
	}

	buildVervImage(t, dockerAPI, tag, files)

	createBoundResourceContainer(t, dockerAPI, vervResSvcName, "pg")

	createReq := newVervonomiconRequest(vervResSvcName, tag, "")

	env.CreateSmerd(t, createReq)

	req := &velez_api.GetVervonomicon_Request{ServiceName: vervResSvcName}

	resp, err := env.Custom.ServiceApiImpl.GetVervonomicon(ctx, req)
	require.NoError(t, err)

	statuses := make(map[string]velez_api.ResourceConnectionStatus, len(resp.GetResourceStatuses()))
	for _, rr := range resp.GetResourceStatuses() {
		statuses[rr.GetName()] = rr.GetStatus()
	}

	require.Equal(t, velez_api.ResourceConnectionStatus_RESOURCE_CONNECTION_STATUS_ALREADY_CONNECTED,
		statuses["pg"], "a resource with an existing service_resources binding must be already_connected")
	require.Equal(t, velez_api.ResourceConnectionStatus_RESOURCE_CONNECTION_STATUS_MUST_PROVISION,
		statuses["cache"], "a resource with no binding and no live matreshka connection must need provisioning")
}

// Test_NoDescriptorIsNotAnError covers the spec's explicit guarantee: a
// deployed image that simply has no /verv directory is not an error -
// GetVervonomicon returns a clean, empty response.
func (s *VervonomiconSuite) Test_NoDescriptorIsNotAnError() {
	t := s.T()
	t.Parallel()

	env := s.plane.NewEnvironment(t)
	ctx := t.Context()

	createReq := newVervonomiconRequest(vervNoDescSvcName, HelloWorldAppImage, "")

	env.CreateSmerd(t, createReq)

	req := &velez_api.GetVervonomicon_Request{ServiceName: vervNoDescSvcName}

	resp, err := env.Custom.ServiceApiImpl.GetVervonomicon(ctx, req)
	require.NoError(t, err)
	require.Empty(t, resp.GetRaw())
	require.Empty(t, resp.GetResolvedYaml())
	require.Equal(t, velez_api.VervonomiconSource_VERVONOMICON_SOURCE_UNSPECIFIED, resp.GetSource())
	require.Empty(t, resp.GetResourceStatuses())
}

// Test_MalformedDescriptorIsAnError covers the flip side: a present
// vervonomicon.yaml this Velez cannot parse (an unrecognised major version)
// IS an error, distinct from ErrNoDescriptor.
func (s *VervonomiconSuite) Test_MalformedDescriptorIsAnError() {
	t := s.T()
	t.Parallel()

	env := s.plane.NewEnvironment(t)
	ctx := t.Context()
	dockerAPI := env.Custom.NodeClients.Docker().Client()

	tag := vervBadSvcName + ":verv"

	files := map[string]string{
		vervIndexFile: "version: \"2\"\nservice:\n  name: " + vervBadSvcName + "\n",
	}

	buildVervImage(t, dockerAPI, tag, files)

	createReq := newVervonomiconRequest(vervBadSvcName, tag, "")

	env.CreateSmerd(t, createReq)

	req := &velez_api.GetVervonomicon_Request{ServiceName: vervBadSvcName}

	_, err := env.Custom.ServiceApiImpl.GetVervonomicon(ctx, req)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unrecognised vervonomicon version")
}

func Test_Vervonomicon(t *testing.T) {
	t.Parallel()
	RunPlaneSuite(t, Planes, func(plane Plane) suite.TestingSuite {
		return &VervonomiconSuite{plane: plane}
	})
}
