package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

// UpgradeSmerdSuite exercises the "upgrade_smerd" job through the real
// UpgradeSmerd RPC (internal/transport/velez_api_impl/smerd_upgrade.go),
// which enqueues onto the jobs engine and blocks on Watch - the same
// Enqueue/Watch glue AssembleConfigSuite covers for assemble_config. The
// job's own 15-step orchestration already has unit-level coverage with
// fakes (internal/jobs/upgrade_smerd_test.go); this suite is what proves the
// real RPC handler drives it correctly against real Docker.
//
// Test_UpgradeSmerd_Matrix iterates the package-wide Plane matrix
// (matrix_test.go) instead of hand-rolling a default-vs-suffixed-environment
// pair of subtests: it carries an extra suite-local `environment` axis (the
// non-default, suffixed environment UpgradeSmerd also has to work against)
// exactly the way suite_container_runtime_test.go's containerRuntimeTestCase
// does. Only the single-node/docker Plane is wired today; a cluster cell
// joins Planes rather than this suite.
type UpgradeSmerdSuite struct {
	suite.Suite
}

// upgradeSmerdTestCase only carries what's specific to this suite's cases -
// which Plane, and which environment (empty => the default/PROD environment;
// non-empty => a suffixed, non-default environment seeded via
// WithEnvironments). The state/backend axes live on Plane itself.
type upgradeSmerdTestCase struct {
	plane       Plane
	environment string
	smerdName   string
}

const (
	// Fixed, suite-unique container names: GetServiceName(t) on a subtest of
	// this matrix contains slashes (from the Plane name), and container names
	// double as Docker hostnames (64-char cap), so each case uses its own
	// short name the same way containerRuntimeMatrix does.
	upgradeDefaultName = "e2e_upgrade_default"

	// upgradeSuffixedName / upgradeSuffixedEnv: a non-default environment.
	// Named environments seeded via WithEnvironments (see environments.NewStatic)
	// always get their own name as their Docker suffix - WithContainerSuffix
	// only ever configures the DEFAULT/PROD environment's suffix, so it has no
	// effect here and is deliberately not used: the real Docker name this
	// case's container carries is "<name>_E2EUPGSTAGE".
	upgradeSuffixedName = "e2e_upgrade_suffixed"
	upgradeSuffixedEnv  = "E2EUPGSTAGE"
)

var upgradeSmerdMatrix = []upgradeSmerdTestCase{
	{plane: Planes[0], environment: "", smerdName: upgradeDefaultName},
	{plane: Planes[0], environment: upgradeSuffixedEnv, smerdName: upgradeSuffixedName},
	// A cluster/docker case joins once Planes grows that cell - see
	// matrix_test.go.
}

// Test_UpgradeSmerd_Matrix proves UpgradeSmerd works end to end for every
// matrix cell: create a hello_world, upgrade it to a newer tag, then assert
// the upgraded container reports the new image, stays running, keeps its
// virtual name, keeps its verv-stamped labels, and - straight off the Docker
// daemon - still carries the environment suffix in its real container name.
//
// Left serial: rows share fixed container names, so parallel rows (or a
// parallel parent racing another suite) would collide in Docker's global
// namespace. Two short cases, cheap enough sequentially.
func (s *UpgradeSmerdSuite) Test_UpgradeSmerd_Matrix() {
	for _, tc := range upgradeSmerdMatrix {
		s.T().Run(tc.plane.Name+"/"+caseEnvName(tc.environment), func(t *testing.T) {
			runUpgradeSmerdCase(t, tc)
		})
	}
}

// caseEnvName labels the subtest by its environment axis - "default" for the
// PROD/unsuffixed case, the raw suffix otherwise.
func caseEnvName(environment string) string {
	if environment == "" {
		return "default"
	}

	return environment
}

// runUpgradeSmerdCase stands up the fixture matching the case's Plane. Only
// single-node/docker is wired - Plane.NewEnvironment fails loudly for anything
// else rather than silently building the wrong fixture, so adding a matrix row
// without its fixture is impossible to miss.
func runUpgradeSmerdCase(t *testing.T, tc upgradeSmerdTestCase) {
	t.Helper()

	var opts []TestEnvOpt

	if tc.environment != "" {
		opts = append(opts, WithEnvironments([]string{tc.environment}))
	}

	env := tc.plane.NewEnvironment(t, opts...)

	createReq := &velez_api.CreateSmerd_Request{
		Name:         tc.smerdName,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
		Environment:  tc.environment,
	}
	created := env.CreateSmerd(t, createReq)
	require.Equal(t, velez_api.Smerd_running, created.GetStatus())

	upgradeReq := &velez_api.UpgradeSmerd_Request{
		Name:        tc.smerdName,
		Image:       helloWorldImageV0015,
		Environment: tc.environment,
	}
	_, err := env.Custom.ApiGrpcImpl.UpgradeSmerd(t.Context(), upgradeReq)
	require.NoError(t, err)

	listReq := &velez_api.ListSmerds_Request{
		Environment: tc.environment,
		Label:       map[string]string{testCaseNameLabel: t.Name()},
	}
	resp := env.ListSmerds(t, t.Context(), listReq)
	require.Len(t, resp.GetSmerds(), 1,
		"expected exactly one container named %q in environment %q after upgrade",
		tc.smerdName, caseEnvName(tc.environment))

	upgraded := resp.GetSmerds()[0]
	require.Equal(t, helloWorldImageV0015, upgraded.GetImageName())
	require.Equal(t, velez_api.Smerd_running, upgraded.GetStatus())
	// Smerd.Name is always the virtual/logical name, never the suffixed Docker
	// container name - see docs/container_runtimes/interface_design.md and
	// suite_container_runtime_test.go's identical assertion.
	require.Equal(t, tc.smerdName, upgraded.GetName(),
		"Smerd.Name must stay the bare/virtual name - no suffix leaking")

	// The verv-stamped labels must survive the rename/recreate the upgrade job
	// runs. SuffixLabel additionally proves the upgraded container stayed in
	// its own environment. (MATRESHKA_CONFIG_ENABLED is deliberately not
	// asserted here - the upgrade job re-stamps it "true" regardless of the
	// original create's IgnoreConfig, which is out of scope for this suite.)
	wantLabels := map[string]string{
		labels.CreatedWithVelezLabel: labelValueTrue,
	}
	if tc.environment != "" {
		wantLabels[labels.SuffixLabel] = tc.environment
	}

	checkVervLabels(t, upgraded.GetLabels(), wantLabels)

	// The checks above all go through ListSmerds/InspectSmerd, which filter by
	// labels.SuffixLabel and return the virtual name via virtualName()'s
	// strings.TrimSuffix - so they'd stay green even if renameContainerJob's
	// raw dockerAPI.ContainerRename calls (internal/jobs/upgrade_smerd.go)
	// never carried the suffix through to the real Docker container name.
	// Inspect the daemon directly, the same way Test_ContainerRuntime_Matrix
	// does. expectedContainerName is a no-op for the empty (default) suffix.
	expectedRealName := expectedContainerName(tc.smerdName, tc.environment)

	dockerClient := env.Custom.NodeClients.Docker().Client()

	inspected, err := dockerClient.ContainerInspect(t.Context(), expectedRealName)
	require.NoError(t, err,
		"expected the upgraded container's real Docker name %q to carry the environment suffix - "+
			"renameContainerJob.Do calls the raw dockerAPI.ContainerRename with a bare, un-suffixed "+
			"name (see internal/jobs/upgrade_smerd.go)", expectedRealName)
	require.Equal(t, "/"+expectedRealName, inspected.Name)
}

func (s *UpgradeSmerdSuite) Test_UpgradeSmerd_NonExistentContainer_Fails() {
	t := s.T()

	serviceName := GetServiceName(t)
	env := Planes[0].NewEnvironment(t)

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
