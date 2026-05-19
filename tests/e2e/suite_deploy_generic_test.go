package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	rtb "go.redsock.ru/toolbox"
	"go.vervstack.ru/matreshka/pkg/matreshka_api"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/tests/config_mocks"
)

type DeploySuite struct {
	suite.Suite
}

func (s *DeploySuite) Test_GenericApi() {
	t := s.T()
	t.Parallel()

	type testCase struct {
		appVervConfig *matreshka_api.StoreConfig_Request

		req      *velez_api.CreateSmerd_Request
		expected *velez_api.Smerd
	}

	testCases := map[string]struct {
		assemble func(t *testing.T) testCase
	}{
		"loki": {
			assemble: func(t *testing.T) (tc testCase) {
				serviceName := GetServiceName(t)

				tc.appVervConfig = &matreshka_api.StoreConfig_Request{
					Format:     matreshka_api.Format_yaml,
					ConfigName: serviceName,
					Config:     config_mocks.Loki,
				}

				tc.req = &velez_api.CreateSmerd_Request{
					Name:      serviceName,
					ImageName: "grafana/loki:main-bc418c4",
					Settings: &velez_api.Container_Settings{
						Network: []*velez_api.NetworkBind{
							{
								NetworkName: "redsockru",
							},
						},
					},
					Restart: &velez_api.RestartPolicy{
						Type: velez_api.RestartPolicyType_always,
					},
					Config: &velez_api.CreateSmerd_Request_Verv{
						Verv: &velez_api.MatreshkaConfigSpec{
							ConfigName:    nil,
							ConfigVersion: nil,
							ConfigFormat:  nil,
							SystemPath:    rtb.ToPtr("/etc/loki/local-config.yaml"),
						},
					},
				}

				tc.expected = &velez_api.Smerd{
					Name:      serviceName,
					ImageName: "grafana/loki:main-bc418c4",
					Ports:     nil,
					Volumes:   nil,
					Status:    velez_api.Smerd_running,
					CreatedAt: nil,
					Networks: []*velez_api.NetworkBind{
						{
							NetworkName: "bridge",
						},
						{
							NetworkName: "redsockru",
							Aliases:     []string{serviceName},
						},
					},
					Labels: map[string]string{
						labels.CreatedWithVelezLabel: "true",
						labels.ComposeGroupLabel:     serviceName,
					},
					Env: map[string]string{
						"VERV_NAME": serviceName,

						"PATH":          "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/busybox",
						"SSL_CERT_FILE": "/etc/ssl/certs/ca-certificates.crt",
					},
				}

				return tc
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()

			tcArgs := tc.assemble(t)

			var opts []TestEnvOpt
			if tcArgs.appVervConfig != nil {
				opts = append(opts, WithMatreshka())
			}

			env := NewEnvironment(t, opts...)

			if tcArgs.appVervConfig != nil {
				_, err := env.Custom.ClusterClients.Configurator().
					StoreConfig(ctx, tcArgs.appVervConfig)
				require.NoError(t, err)
			}

			deployedSmerd := env.CreateSmerd(t, tcArgs.req)

			AssertSmerds(t, tcArgs.expected, deployedSmerd)
		})
	}
}

func Test_Deploy(t *testing.T) {
	suite.Run(t, new(DeploySuite))
}
