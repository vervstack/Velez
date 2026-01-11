package tests

import (
	"testing"

	"github.com/stretchr/testify/require"
	rtb "go.redsock.ru/toolbox"
	"go.vervstack.ru/matreshka/pkg/matreshka_api"

	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/pkg/velez_api"
	"go.vervstack.ru/Velez/tests/config_mocks"
)

func Test_Deploy_Generic_Api(t *testing.T) {
	t.Parallel()

	type testCase struct {
		appVervConfig *matreshka_api.StoreConfig_Request

		req      *velez_api.CreateSmerd_Request
		expected *velez_api.Smerd
	}

	timeoutSeconds := uint32(5)

	testCases := map[string]struct {
		assemble func(t *testing.T) testCase
	}{
		"postgres": {
			assemble: func(t *testing.T) testCase {
				serviceName := GetServiceName(t)

				tc := testCase{
					req: &velez_api.CreateSmerd_Request{
						Name:      serviceName,
						ImageName: PostgresImage,
						Env: map[string]string{
							"POSTGRES_HOST_AUTH_METHOD": "trust",
						},
						Labels: GetExpectedLabels(t),
						Healthcheck: &velez_api.Container_Healthcheck{
							Command:        rtb.ToPtr("pg_isready -U postgres"),
							IntervalSecond: 2,
							TimeoutSecond:  &timeoutSeconds,
							Retries:        3,
						},
						IgnoreConfig:  true,
						UseImagePorts: true,
					},
					expected: &velez_api.Smerd{
						Name:      serviceName,
						ImageName: PostgresImage,
						Status:    velez_api.Smerd_running,
						Labels:    GetExpectedLabels(t),
						Env: map[string]string{
							"POSTGRES_HOST_AUTH_METHOD": "trust",
							"GOSU_VERSION":              "1.17",
							"LANG":                      "en_US.utf8",
							"PATH":                      "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/lib/postgresql/16/bin",
							"PGDATA":                    "/var/lib/postgresql/data",
							"PG_MAJOR":                  "16",
							"PG_VERSION":                "16.9-1.pgdg120+1",
						},
						Ports: []*velez_api.Port{
							{
								ServicePortNumber: 5432,
								Protocol:          velez_api.Port_tcp,
								ExposedTo:         nil,
							},
						},
						Networks: []*velez_api.NetworkBind{
							{
								NetworkName: "bridge",
								Aliases:     []string{},
							},
							{
								NetworkName: "verv",
								Aliases: []string{
									GetServiceName(t),
								},
							},
						},
					},
				}

				tc.expected.Labels[labels.ComposeGroupLabel] = GetServiceName(t)

				return tc
			},
		},
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
			// TODO Currently this test case checks deployment only.
			// Target for future improvement is to perform some work e.g
			//    - perform sql queries for postgres,
			//    - store and retrieve some data to/from loki
			ctx := t.Context()
			// region Preparations
			tcArgs := tc.assemble(t)
			var opts []TestEnvOpt

			if tcArgs.appVervConfig != nil {
				opts = append(opts, WithMatreshka())
			}

			//endregion
			env := NewEnvironment(t, opts...)

			if tcArgs.appVervConfig != nil {
				_, err := env.Custom.ClusterClients.Configurator().
					StoreConfig(ctx, tcArgs.appVervConfig)
				require.NoError(t, err)
			}

			deployedSmerd, err := env.CreateSmerd(ctx, tcArgs.req)
			require.NoError(t, err)

			AssertSmerds(t, tcArgs.expected, deployedSmerd)
		})
	}
}

func Test_Deploy_Verv_Api(t *testing.T) {
	t.Parallel()

	type testCase struct {
		req        *velez_api.CreateSmerd_Request
		expected   *velez_api.Smerd
		nameSuffix string
	}

	testCases := map[string]struct {
		new func(t *testing.T) (tc testCase)
	}{
		"OK_WITHOUT_CONFIG": {
			new: func(t *testing.T) (tc testCase) {
				serviceName := GetServiceName(t)

				tc.req = &velez_api.CreateSmerd_Request{
					Name:         serviceName,
					ImageName:    HelloWorldAppImage,
					IgnoreConfig: true,
				}

				tc.expected = &velez_api.Smerd{
					Name:      serviceName,
					ImageName: HelloWorldAppImage,
					Status:    velez_api.Smerd_running,
					Labels:    GetExpectedLabels(t),
					Env: map[string]string{
						"PATH": "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
					},
					Networks: []*velez_api.NetworkBind{
						{
							NetworkName: "bridge",
							Aliases:     nil,
						},
					},
				}
				return
			},
		},
		"OK_WITH_HEALTH_CHEKS": {
			new: func(t *testing.T) (tc testCase) {
				serviceName := GetServiceName(t)

				tc.req = &velez_api.CreateSmerd_Request{
					Name:      serviceName,
					ImageName: HelloWorldAppImage,
					Healthcheck: &velez_api.Container_Healthcheck{
						IntervalSecond: 1,
						Retries:        3,
					},
					IgnoreConfig: true,
				}

				tc.expected = &velez_api.Smerd{
					Name:      serviceName,
					ImageName: HelloWorldAppImage,
					Status:    velez_api.Smerd_running,
					Labels:    GetExpectedLabels(t),
					Env: map[string]string{
						"PATH": "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
					},
					Networks: []*velez_api.NetworkBind{
						{
							NetworkName: "bridge",
							Aliases:     nil,
						},
					},
				}
				return
			},
		},
		"OK_WITH_DEFAULT_CONFIG": {
			new: func(t *testing.T) (tc testCase) {
				serviceName := GetServiceName(t)

				tc.req = &velez_api.CreateSmerd_Request{
					Name:      serviceName,
					ImageName: HelloWorldAppImage,
				}

				tc.expected = &velez_api.Smerd{
					Name:      serviceName,
					Status:    velez_api.Smerd_running,
					ImageName: HelloWorldAppImage,
					Labels:    GetExpectedLabels(t),
					Env: map[string]string{
						"PATH":      "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
						"VERV_NAME": GetServiceName(t),
					},
					Networks: []*velez_api.NetworkBind{
						{
							NetworkName: "bridge",
							Aliases:     nil,
						},
					},
				}
				tc.expected.Labels[labels.MatreshkaConfigLabel] = "true"
				return
			},
		},
	}

	for name, tcConstructor := range testCases {
		t.Run(name,
			func(t *testing.T) {
				t.Parallel()
				ctx := t.Context()

				tc := tcConstructor.new(t)

				env := NewEnvironment(t)

				tc.req.Name = GetServiceName(t)

				launchedSmerd, err := env.CreateSmerd(ctx, tc.req)
				require.NoError(t, err)

				AssertSmerds(t, tc.expected, launchedSmerd)
			})
	}
}
