package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.redsock.ru/toolbox"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

// UpgradeSmerdSuite exercises the "upgrade_smerd" job through the real
// UpgradeSmerd RPC (internal/transport/velez_api_impl/smerd_upgrade.go),
// which enqueues onto the jobs engine and blocks on Watch - the same
// Enqueue/Watch glue AssembleConfigSuite covers for assemble_config. The
// job's own 15-step orchestration already has unit-level coverage with
// fakes (internal/jobs/upgrade_smerd_test.go); this suite is what proves the
// real RPC handler drives it correctly against real Docker.
type UpgradeSmerdSuite struct {
	suite.Suite
}

func (s *UpgradeSmerdSuite) Test_UpgradeSmerd_HappyPath() {
	t := s.T()

	serviceName := GetServiceName(t)
	env := NewEnvironment(t)

	createReq := &velez_api.CreateSmerd_Request{
		Name:         serviceName,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
	}
	created := env.CreateSmerd(t, createReq)
	require.Equal(t, velez_api.Smerd_running, created.GetStatus())

	upgradeReq := &velez_api.UpgradeSmerd_Request{
		Name:  serviceName,
		Image: helloWorldImageV0015,
	}
	_, err := env.Custom.ApiGrpcImpl.UpgradeSmerd(t.Context(), upgradeReq)
	require.NoError(t, err)

	listReq := &velez_api.ListSmerds_Request{Name: toolbox.ToPtr(serviceName)}
	resp := env.ListSmerds(t, t.Context(), listReq)
	require.Len(t, resp.GetSmerds(), 1, "expected exactly one container named %q after upgrade", serviceName)

	upgraded := resp.GetSmerds()[0]
	require.Equal(t, helloWorldImageV0015, upgraded.GetImageName())
	require.Equal(t, velez_api.Smerd_running, upgraded.GetStatus())
}

const (
	upgradeSuffixedEnvName = "e2e_upgrade_suffixed"

	// A non-default environment - UpgradeSmerd today only works against the
	// default (PROD, unsuffixed) environment. Named environments seeded via
	// WithEnvironments (see environments.NewStatic) always get their own name
	// as their Docker suffix - WithContainerSuffix only ever configures the
	// DEFAULT/PROD environment's suffix, so it has no effect here and is
	// deliberately not used: the real Docker name this test's container
	// carries is "<name>_E2EUPGSTAGE", not some separately configured value.
	upgradeSuffixedEnv = "E2EUPGSTAGE"
)

// Test_UpgradeSmerd_InSuffixedEnvironment is a RED test proving UpgradeSmerd
// is completely broken for any non-default (suffixed) environment. Root
// causes (see internal/jobs/upgrade_smerd.go and
// internal/transport/velez_api_impl/smerd_upgrade.go):
//
//  1. Impl.UpgradeSmerd validates req.GetEnvironment() but never copies it
//     onto the initialContext.UpgradeRequest it builds - Environment is
//     dropped before the request ever reaches the jobs engine.
//  2. captureOldContainerJob.Do builds a CreateSmerd_Request without ever
//     setting Environment on it, even though every downstream step
//     (port locking, network creation, etc.) reads Environment off that
//     request.
//  3. checkSelfUpgradeJob/captureOldContainerJob look up "the current
//     container" via containerService.InspectSmerd(ctx, bareName) - a raw,
//     suffix-blind ContainerInspect. In a suffixed environment the real
//     Docker container is named "name_<suffix>"
//     (container_runtime.labelBasedRuntime.containerName), so this 404s
//     immediately, before any upgrade logic runs.
//
// It mirrors Test_UpgradeSmerd_HappyPath, but stands up a suffixed,
// non-default environment first (same fixture shape as
// suite_environments_test.go's Test_TwoEnvironments_AreListScoped and
// suite_container_runtime_test.go's
// Test_ContainerRuntime_ListContainers_ScopesToEnvironment).
func (s *UpgradeSmerdSuite) Test_UpgradeSmerd_InSuffixedEnvironment() {
	t := s.T()

	env := NewEnvironment(t, WithEnvironments([]string{upgradeSuffixedEnv}))

	createReq := &velez_api.CreateSmerd_Request{
		Name:         upgradeSuffixedEnvName,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
		Environment:  upgradeSuffixedEnv,
	}
	created := env.CreateSmerd(t, createReq)
	require.Equal(t, velez_api.Smerd_running, created.GetStatus())

	upgradeReq := &velez_api.UpgradeSmerd_Request{
		Name:        upgradeSuffixedEnvName,
		Image:       helloWorldImageV0015,
		Environment: upgradeSuffixedEnv,
	}
	_, err := env.Custom.ApiGrpcImpl.UpgradeSmerd(t.Context(), upgradeReq)
	require.NoError(t, err)

	listReq := &velez_api.ListSmerds_Request{
		Environment: upgradeSuffixedEnv,
		Label:       map[string]string{testCaseNameLabel: t.Name()},
	}
	resp := env.ListSmerds(t, t.Context(), listReq)
	require.Len(t, resp.GetSmerds(), 1,
		"expected exactly one container named %q in environment %q after upgrade",
		upgradeSuffixedEnvName, upgradeSuffixedEnv)

	upgraded := resp.GetSmerds()[0]
	require.Equal(t, helloWorldImageV0015, upgraded.GetImageName())
	require.Equal(t, velez_api.Smerd_running, upgraded.GetStatus())
	// Smerd.Name is always the virtual/logical name, never the suffixed
	// Docker container name - see
	// docs/container_runtimes/interface_design.md and
	// suite_container_runtime_test.go's identical assertion.
	require.Equal(t, upgradeSuffixedEnvName, upgraded.GetName(),
		"Smerd.Name must stay the bare/virtual name - no suffix leaking")

	// The checks above only go through ListSmerds/InspectSmerd, which filter
	// by labels.SuffixLabel (untouched by renaming) and return the
	// virtual/bare name via virtualName()'s strings.TrimSuffix - a harmless
	// no-op on an already-unsuffixed real name. They would still pass even if
	// renameContainerJob's raw dockerAPI.ContainerRename calls (see
	// internal/jobs/upgrade_smerd.go) never carried the suffix through to the
	// real Docker container name. So inspect the daemon directly, the same
	// way Test_ContainerRuntime_Matrix does, to prove the *real* Docker name
	// of the fully-upgraded container still carries the environment suffix.
	dockerClient := env.Custom.NodeClients.Docker().Client()
	expectedRealName := expectedContainerName(upgradeSuffixedEnvName, upgradeSuffixedEnv)

	inspected, err := dockerClient.ContainerInspect(t.Context(), expectedRealName)
	require.NoError(t, err,
		"expected the upgraded container's real Docker name %q to carry the environment suffix - "+
			"renameContainerJob.Do calls the raw dockerAPI.ContainerRename with a bare, un-suffixed "+
			"name (see internal/jobs/upgrade_smerd.go), so the real container likely ended up named "+
			"%q instead", expectedRealName, upgradeSuffixedEnvName)
	require.Equal(t, "/"+expectedRealName, inspected.Name)
}

func (s *UpgradeSmerdSuite) Test_UpgradeSmerd_NonExistentContainer_Fails() {
	t := s.T()

	serviceName := GetServiceName(t)
	env := NewEnvironment(t)

	upgradeReq := &velez_api.UpgradeSmerd_Request{
		Name:  serviceName,
		Image: helloWorldImageV0015,
	}
	_, err := env.Custom.ApiGrpcImpl.UpgradeSmerd(t.Context(), upgradeReq)
	require.Error(t, err)
}

func Test_UpgradeSmerd(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(UpgradeSmerdSuite))
}
