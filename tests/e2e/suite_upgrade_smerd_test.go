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
	suite.Run(t, new(UpgradeSmerdSuite))
}
