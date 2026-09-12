//go:build e2e_full

package e2e

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

// VervConfigSuite covers config delivery into a running smerd: a verv-classified
// deploy that injects VERV_NAME, a Plain file-config mounted verbatim, and the
// restart policy landing on the container HostConfig. It lives in package e2e
// alongside the shared DinD / matreshka fixtures.
type VervConfigSuite struct {
	suite.Suite

	plane Plane
	ctx   context.Context
}

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

// Test_VervConfig_RenderedEnv deploys hello_world with Verv set and
// IgnoreConfig unset so fetchSmerdConfigJob.doVerv -> setEnv runs, and asserts
// the reachable verv-config invariants: the image is classified as verv
// (MatreshkaConfigLabel true) and VERV_NAME is injected. A Plain mount is
// attached alongside and asserted to land in the container.
//
// TODO(phase-1): assert real rendered matreshka config values once the
// verv://matreshka gRPC resolver can be made to resolve inside the e2e
// harness. WithMatreshka() turns that resolver path on, but it currently
// "produces zero addresses" for both the raw configurator client and the
// fetch_config job, so a real config pre-seed + !IgnoreConfig deploy against
// the shared matreshka cannot be exercised here yet.
func (s *VervConfigSuite) Test_VervConfig_RenderedEnv() {
	t := s.T()

	env := s.plane.NewEnvironment(t)

	const plainPath = "/tmp/verv_rendered_test.yaml"

	plainContent := []byte("seeded: true\n")

	req := newVervConfigRenderedEnvRequest(GetServiceName(t), plainPath, plainContent)

	smerd := env.CreateSmerd(t, req)

	require.Equal(t, velez_api.Smerd_running.String(), smerd.GetStatus().String())
	require.Equal(t, labelValueTrue, smerd.GetLabels()[labels.MatreshkaConfigLabel])
	require.Equal(t, smerd.GetName(), smerd.GetEnv()["VERV_NAME"])

	dockerClient := env.Custom.NodeClients.Docker().Client()

	got, err := dockerutils.ReadFromContainer(s.ctx, dockerClient, smerd.GetUuid(), plainPath)
	require.NoError(t, err)
	require.Equal(t, string(plainContent), string(got))
}

// Test_VervConfig_PlainFileMounted deploys with a Plain file-config and asserts
// the file is present inside the running container with the exact bytes given.
func (s *VervConfigSuite) Test_VervConfig_PlainFileMounted() {
	t := s.T()

	env := s.plane.NewEnvironment(t)

	// NOTE(phase-1): create_smerd's copyToContainerJob copies via
	// dockerutils.WriteToContainer with no "mkdir -p" of the parent dir
	// (unlike copy_to_volume.go's copyFileJob), so a Plain path under a
	// directory absent from the image fails the whole task at
	// copy_to_container. /tmp exists in the image, so the mount lands.
	// The missing-parent-dir gap is reported, not fixed in this test-only phase.
	const filePath = "/tmp/verv_test.yaml"

	wantContent := []byte("key: value\n")

	req := newVervConfigPlainFileRequest(GetServiceName(t), filePath, wantContent)

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

	env := s.plane.NewEnvironment(t)

	req := newVervConfigRestartAlwaysRequest(GetServiceName(t))

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
