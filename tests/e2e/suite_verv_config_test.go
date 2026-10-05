//go:build e2e_full

package e2e

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"go.vervstack.ru/Velez/internal/api/clients/matreshka/pkg/matreshka_api"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/utils/configutils"
)

// VervConfigSuite covers config delivery into a running smerd: a verv-classified
// deploy that injects VERV_NAME, a Plain file-config mounted verbatim, and the
// restart policy landing on the container HostConfig. It lives in package e2e
// alongside the shared DinD / matreshka fixtures.
//
// Container names double as Docker hostnames, capped at 64 characters -
// GetServiceName(t) on a RunPlaneSuite subtest blows past that (see
// ClusterLifecycleSuite's doc comment in suite_api_deploy_test.go), so these
// tests use their own short, suite-unique names instead.
type VervConfigSuite struct {
	suite.Suite

	plane Plane
	ctx   context.Context
}

const (
	vervConfigSeededKey   = "E2E_VERV_SEEDED"
	vervConfigSeededValue = "from-matreshka"
	vervConfigSecondKey   = "E2E_VERV_SECOND"
	vervConfigSecondValue = "also-from-matreshka"

	vervConfigRenderedEnvName   = "e2e_vervconfig_renderedenv"
	vervConfigPlainFileName     = "e2e_vervconfig_plainfile"
	vervConfigRestartPolicyName = "e2e_vervconfig_restartpolicy"
)

func (s *VervConfigSuite) SetupSuite() {
	s.ctx = context.Background()
}

func newVervConfigRenderedEnvRequest(name, plainPath string, plainContent []byte) *velez_api.CreateSmerd_Request {
	return &velez_api.CreateSmerd_Request{
		Name:      name,
		ImageName: HelloWorldAppImage,
		Verv:      &velez_api.MatreshkaConfigSpec{},
		Plain: []*velez_api.FileConfig{
			{Path: plainPath, Content: plainContent},
		},
	}
}

func newVervConfigSeededEnv() map[string]string {
	return map[string]string{
		vervConfigSeededKey: vervConfigSeededValue,
		vervConfigSecondKey: vervConfigSecondValue,
	}
}

// seedMatreshkaEnv writes env into the shared matreshka as a kv config the
// same way configresolver does: SaveConfig can only fill an empty config from
// a value with no '=' in it, so one key is seeded and all of them patched.
func seedMatreshkaEnv(t *testing.T, env *TestEnvironment, configName string, values map[string]string) {
	t.Helper()

	ctx := t.Context()
	configurator := env.Custom.ClusterClients.Configurator()

	createReq := &matreshka_api.CreateConfig_Request{
		ConfigName: configName,
		ConfigType: matreshka_api.ConfigType_kv,
	}

	_, err := configurator.CreateConfig(ctx, createReq)
	require.NoError(t, err)

	t.Cleanup(func() {
		deleteReq := &matreshka_api.DeleteConfig_Request{ConfigName: configName}

		_, _ = configurator.DeleteConfig(context.Background(), deleteReq)
	})

	patches := make([]*matreshka_api.Patch, 0, len(values))

	var seedKey string

	for key, value := range values {
		seedKey = key

		update := &matreshka_api.Patch_UpdateValue{UpdateValue: value}

		patches = append(patches, &matreshka_api.Patch{FieldName: key, Patch: update})
	}

	saveReq := &matreshka_api.SaveConfig_Request{
		Format:     matreshka_api.Format_env,
		ConfigName: configName,
		Config:     []byte(seedKey + "=seed\n"),
	}

	_, err = configurator.SaveConfig(ctx, saveReq)
	require.NoError(t, err)

	patchReq := &matreshka_api.PatchConfig_Request{ConfigName: configName, Patches: patches}

	_, err = configurator.PatchConfig(ctx, patchReq)
	require.NoError(t, err)
}

func newVervConfigPlainFileRequest(name, filePath string, wantContent []byte) *velez_api.CreateSmerd_Request {
	return &velez_api.CreateSmerd_Request{
		Name:         name,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
		Plain: []*velez_api.FileConfig{
			{Path: filePath, Content: wantContent},
		},
	}
}

func newVervConfigRestartAlwaysRequest(name string) *velez_api.CreateSmerd_Request {
	return &velez_api.CreateSmerd_Request{
		Name:         name,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
		Restart: &velez_api.RestartPolicy{
			Type: velez_api.RestartPolicyType_always,
		},
	}
}

// Test_VervConfig_RenderedEnv pre-seeds a real matreshka config, deploys
// hello_world with Verv set and IgnoreConfig unset so fetchSmerdConfigJob.doVerv
// -> setEnv pulls it, and asserts the image is classified as verv
// (MatreshkaConfigLabel true), VERV_NAME is injected and the seeded values
// reach the running container. A Plain mount is attached alongside and
// asserted to land in the container.
func (s *VervConfigSuite) Test_VervConfig_RenderedEnv() {
	t := s.T()
	t.Parallel()

	env := s.plane.NewEnvironment(t, WithMatreshka())

	const plainPath = "/tmp/verv_rendered_test.yaml"

	configName := configutils.AppendPrefix(matreshka_api.ConfigType_verv, vervConfigRenderedEnvName)
	seedMatreshkaEnv(t, env, configName, newVervConfigSeededEnv())

	plainContent := []byte("seeded: true\n")

	req := newVervConfigRenderedEnvRequest(vervConfigRenderedEnvName, plainPath, plainContent)

	smerd := env.CreateSmerd(t, req)

	require.Equal(t, velez_api.Smerd_running.String(), smerd.GetStatus().String())
	require.Equal(t, labelValueTrue, smerd.GetLabels()[labels.MatreshkaConfigLabel])
	require.Equal(t, smerd.GetName(), smerd.GetEnv()["VERV_NAME"])
	require.Equal(t, vervConfigSeededValue, smerd.GetEnv()[vervConfigSeededKey])
	require.Equal(t, vervConfigSecondValue, smerd.GetEnv()[vervConfigSecondKey])

	dockerClient := env.Custom.NodeClients.Docker().Client()

	got, err := dockerutils.ReadFromContainer(s.ctx, dockerClient, smerd.GetUuid(), plainPath)
	require.NoError(t, err)
	require.Equal(t, string(plainContent), string(got))
}

// Test_VervConfig_PlainFileMounted deploys with a Plain file-config and asserts
// the file is present inside the running container with the exact bytes given.
func (s *VervConfigSuite) Test_VervConfig_PlainFileMounted() {
	t := s.T()
	t.Parallel()

	env := s.plane.NewEnvironment(t)

	// NOTE(phase-1): create_smerd's copyToContainerJob copies via
	// dockerutils.WriteToContainer with no "mkdir -p" of the parent dir
	// (unlike copy_to_volume.go's copyFileJob), so a Plain path under a
	// directory absent from the image fails the whole task at
	// copy_to_container. /tmp exists in the image, so the mount lands.
	// The missing-parent-dir gap is reported, not fixed in this test-only phase.
	const filePath = "/tmp/verv_test.yaml"

	wantContent := []byte("key: value\n")

	req := newVervConfigPlainFileRequest(vervConfigPlainFileName, filePath, wantContent)

	smerd := env.CreateSmerd(t, req)

	require.Equal(t, velez_api.Smerd_running.String(), smerd.GetStatus().String())

	dockerClient := env.Custom.NodeClients.Docker().Client()

	got, err := dockerutils.ReadFromContainer(s.ctx, dockerClient, smerd.GetUuid(), filePath)
	require.NoError(t, err)
	require.Equal(t, string(wantContent), string(got))
}

// Test_VervConfig_RestartPolicyApplied ports the restart-policy intent of the
// skipped Test_StatelessMode_Loki: deploy with RestartPolicyType_always and
// assert a policy actually reached the container HostConfig.
func (s *VervConfigSuite) Test_VervConfig_RestartPolicyApplied() {
	t := s.T()
	t.Parallel()

	env := s.plane.NewEnvironment(t)

	req := newVervConfigRestartAlwaysRequest(vervConfigRestartPolicyName)

	smerd := env.CreateSmerd(t, req)

	dockerClient := env.Custom.NodeClients.Docker().Client()

	info, err := dockerClient.ContainerInspect(s.ctx, smerd.GetUuid())
	require.NoError(t, err)

	rp := info.HostConfig.RestartPolicy
	require.NotEmpty(t, string(rp.Name), "a restart policy must be applied to the container")

	// NOTE(phase-1): parser.FromRestart maps RestartPolicyType_always onto
	// docker's "on-failure" (MaximumRetryCount 3), not "always". This asserts
	// the real current behaviour; the always->on-failure mapping is reported
	// as a product gap, not fixed in this test-only phase.
	require.EqualValues(t, "on-failure", rp.Name)
	require.Equal(t, 3, rp.MaximumRetryCount)
}

func Test_VervConfig(t *testing.T) {
	t.Parallel()
	RunPlaneSuite(t, Planes, func(plane Plane) suite.TestingSuite {
		return &VervConfigSuite{plane: plane}
	})
}
