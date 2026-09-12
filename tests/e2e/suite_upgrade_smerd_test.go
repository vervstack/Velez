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

// Test_UpgradeSmerd_Matrix upgrades a smerd's image in both the default (PROD)
// environment and a suffixed one, asserting the new image, labels, and real
// container name land correctly in each. Serial: rows share fixed container
// names, so they can't run in parallel with each other.
func (s *UpgradeSmerdSuite) Test_UpgradeSmerd_Matrix() {
	cases := []upgradeSmerdTestCase{
		{environment: "", smerdName: upgradeDefaultName},
		{environment: upgradeSuffixedEnv, smerdName: upgradeSuffixedName},
	}

	for _, tc := range cases {
		s.T().Run(caseEnvName(tc.environment), func(t *testing.T) {
			runUpgradeSmerdCase(t, s.plane, tc)
		})
	}
}

func caseEnvName(environment string) string {
	if environment == "" {
		return "default"
	}

	return environment
}

func runUpgradeSmerdCase(t *testing.T, plane Plane, tc upgradeSmerdTestCase) {
	t.Helper()

	var opts []TestEnvOpt

	if tc.environment != "" {
		opts = append(opts, WithEnvironments([]string{tc.environment}))
	}

	env := plane.NewEnvironment(t, opts...)

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
	require.Len(t, resp.GetSmerds(), 1, "exactly one smerd after upgrade")

	upgraded := resp.GetSmerds()[0]
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

	env := s.plane.NewEnvironment(t)

	upgradeReq := &velez_api.UpgradeSmerd_Request{
		Name:  GetServiceName(t),
		Image: helloWorldImageV0015,
	}
	_, err := env.Custom.ApiGrpcImpl.UpgradeSmerd(t.Context(), upgradeReq)
	require.Error(t, err)
}

func Test_UpgradeSmerd(t *testing.T) {
	t.Parallel()
	RunPlaneSuite(t, Planes, func(plane Plane) suite.TestingSuite {
		return &UpgradeSmerdSuite{plane: plane}
	})
}
