//go:build e2e_full

package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

// UpgradeSmerdSuite drives the real UpgradeSmerd RPC - which enqueues the
// upgrade_smerd job and blocks on Watch - against real Docker.
type UpgradeSmerdSuite struct {
	suite.Suite

	plane Plane
	env   *TestEnvironment
}

type upgradeSmerdTestCase struct {
	environment string // "" => default/PROD environment
	smerdName   string
}

const (
	upgradeDefaultName  = "e2e_upgrade_default"
	upgradeSuffixedName = "e2e_upgrade_suffixed"
	upgradeSuffixedEnv  = "E2EUPGSTAGE"
)

// SetupTest builds the plain, no-options environment used by
// Test_UpgradeSmerd_NonExistentContainer_Fails. Test_UpgradeSmerd_Matrix needs
// a differently-configured environment per row (a suffixed environment for
// one of its two cases), so it builds its own instead of using s.env.
func (s *UpgradeSmerdSuite) SetupTest() {
	s.env = s.plane.NewEnvironment(s.T())
}

// Serial: rows share fixed container names.
func (s *UpgradeSmerdSuite) Test_UpgradeSmerd_Matrix() {
	cases := []upgradeSmerdTestCase{
		{environment: "", smerdName: upgradeDefaultName},
		{environment: upgradeSuffixedEnv, smerdName: upgradeSuffixedName},
	}

	for _, tc := range cases {
		s.T().Run(caseEnvName(tc.environment), func(t *testing.T) {
			var opts []TestEnvOpt

			if tc.environment != "" {
				opts = append(opts, WithEnvironments([]string{tc.environment}))
			}

			env := s.plane.NewEnvironment(t, opts...)

			created := env.CreateSmerd(t, newCreateSmerdRequest(tc))
			require.Equal(t, velez_api.Smerd_running, created.GetStatus())

			_, err := env.Custom.ApiGrpcImpl.UpgradeSmerd(t.Context(), newUpgradeSmerdRequest(tc))
			require.NoError(t, err)

			upgraded := s.awaitUpgraded(t, env, tc)

			s.assertUpgraded(t, env, tc, upgraded)
		})
	}
}

func caseEnvName(environment string) string {
	if environment == "" {
		return "default"
	}

	return environment
}

func newCreateSmerdRequest(tc upgradeSmerdTestCase) *velez_api.CreateSmerd_Request {
	return &velez_api.CreateSmerd_Request{
		Name:         tc.smerdName,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
		Environment:  tc.environment,
	}
}

func newUpgradeSmerdRequest(tc upgradeSmerdTestCase) *velez_api.UpgradeSmerd_Request {
	return &velez_api.UpgradeSmerd_Request{
		Name:        tc.smerdName,
		Image:       helloWorldImageV0015,
		Environment: tc.environment,
	}
}

func newUpgradedListRequest(t *testing.T, tc upgradeSmerdTestCase) *velez_api.ListSmerds_Request {
	t.Helper()

	return &velez_api.ListSmerds_Request{
		Environment: tc.environment,
		Label:       map[string]string{testCaseNameLabel: t.Name()},
	}
}

func (s *UpgradeSmerdSuite) awaitUpgraded(
	t *testing.T,
	env *TestEnvironment,
	tc upgradeSmerdTestCase,
) *velez_api.Smerd {
	t.Helper()

	resp := env.ListSmerds(t, t.Context(), newUpgradedListRequest(t, tc))
	require.Len(t, resp.GetSmerds(), 1, "exactly one smerd after upgrade")

	return resp.GetSmerds()[0]
}

// assertUpgraded validates the RPC-visible smerd (image, status, name,
// labels) and, separately, that the real Docker container behind it kept its
// environment-suffixed name.
func (s *UpgradeSmerdSuite) assertUpgraded(
	t *testing.T,
	env *TestEnvironment,
	tc upgradeSmerdTestCase,
	upgraded *velez_api.Smerd,
) {
	t.Helper()

	require.Equal(t, helloWorldImageV0015, upgraded.GetImageName())
	require.Equal(t, velez_api.Smerd_running, upgraded.GetStatus())
	require.Equal(t, tc.smerdName, upgraded.GetName())

	// upgrade_smerd re-stamps MATRESHKA_CONFIG_ENABLED "true" regardless of
	// the create's IgnoreConfig, so it is not asserted here.
	wantLabels := map[string]string{labels.CreatedWithVelezLabel: labelValueTrue}

	if tc.environment != "" {
		wantLabels[labels.SuffixLabel] = tc.environment
	}

	checkVervLabels(t, upgraded.GetLabels(), wantLabels)

	// ListSmerds/InspectSmerd return the virtual name; check the daemon
	// directly that the real container name kept its environment suffix.
	expectedRealName := expectedContainerName(tc.smerdName, tc.environment)

	dockerClient := env.Custom.NodeClients.Docker().Client()

	inspected, err := dockerClient.ContainerInspect(t.Context(), expectedRealName)
	require.NoError(t, err)
	require.Equal(t, "/"+expectedRealName, inspected.Name)
}

func (s *UpgradeSmerdSuite) Test_UpgradeSmerd_NonExistentContainer_Fails() {
	t := s.T()

	upgradeReq := &velez_api.UpgradeSmerd_Request{
		Name:  GetServiceName(t),
		Image: helloWorldImageV0015,
	}

	_, err := s.env.Custom.ApiGrpcImpl.UpgradeSmerd(t.Context(), upgradeReq)
	require.Error(t, err)
}

func Test_UpgradeSmerd(t *testing.T) {
	t.Parallel()
	RunPlaneSuite(t, Planes, func(plane Plane) suite.TestingSuite {
		return &UpgradeSmerdSuite{plane: plane}
	})
}
