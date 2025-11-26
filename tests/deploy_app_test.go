package tests

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.redsock.ru/toolbox"
	"go.vervstack.ru/matreshka/pkg/matreshka_api"
	"google.golang.org/protobuf/proto"

	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

const (
	helloWorldAppImage = "godverv/hello_world:v0.0.14"
	postgresImage      = "postgres:16"
)

func Test_DeploySingleVerv(t *testing.T) {
	type testCase struct {
		reqs       []*velez_api.CreateSmerd_Request
		expected   []*velez_api.Smerd
		nameSuffix string
	}

	testCases := map[string]testCase{
		"OK_WITHOUT_CONFIG": {
			reqs: []*velez_api.CreateSmerd_Request{
				{
					ImageName:    helloWorldAppImage,
					IgnoreConfig: true,
				},
			},
			expected: []*velez_api.Smerd{
				{
					ImageName: helloWorldAppImage,
					Status:    velez_api.Smerd_running,
					Labels:    getExpectedLabels(),
					Env: map[string]string{
						"PATH": "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
					},
					Networks: []*velez_api.NetworkBind{
						{
							NetworkName: "bridge",
							Aliases:     nil,
						},
					},
				},
			},
		},
		"OK_WITH_HEALTH_CHEKS": {
			reqs: []*velez_api.CreateSmerd_Request{
				{
					ImageName: helloWorldAppImage,
					Healthcheck: &velez_api.Container_Healthcheck{
						IntervalSecond: 1,
						Retries:        3,
					},
					IgnoreConfig: true,
				},
			},
			expected: []*velez_api.Smerd{
				{
					ImageName: helloWorldAppImage,
					Status:    velez_api.Smerd_running,
					Labels:    getExpectedLabels(),
					Env: map[string]string{
						"PATH": "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
					},
					Networks: []*velez_api.NetworkBind{
						{
							NetworkName: "bridge",
							Aliases:     nil,
						},
					},
				},
			},
		},
		"OK_WITH_DEFAULT_CONFIG": {
			reqs: []*velez_api.CreateSmerd_Request{
				{
					ImageName: helloWorldAppImage,
				},
			},
			expected: []*velez_api.Smerd{
				{
					Status:    velez_api.Smerd_running,
					ImageName: helloWorldAppImage,
					Labels:    getLabelsWithMatreshkaConfig(),
					Env: map[string]string{
						"PATH":      "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
						"VERV_NAME": "Test_DeploySingleVerv_OK_WITH_DEFAULT_CONFIG",
					},
					Networks: []*velez_api.NetworkBind{
						{
							NetworkName: "bridge",
							Aliases:     nil,
						},
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			for idx, createReq := range tc.reqs {
				createReq.Name = getServiceName(t)

				launchedSmerd, err := testEnvironment.createSmerd(ctx, createReq)
				require.NoError(t, err)

				require.NotEmpty(t, launchedSmerd.Uuid)
				launchedSmerd.Uuid = ""

				require.NotEmpty(t, launchedSmerd.CreatedAt)
				launchedSmerd.CreatedAt = nil

				tc.expected[idx].Name = "/" + createReq.Name

				if !proto.Equal(tc.expected[idx], launchedSmerd) {
					require.Equal(t, launchedSmerd, tc.expected[idx])
				}
			}
		})
	}
}

func Test_DeployMultipleVerv(t *testing.T) {
	serviceName := getServiceName(t)
	masterVersion := &velez_api.CreateSmerd_Request{
		Name:      serviceName,
		ImageName: helloWorldAppImage,
	}

	ctx := context.Background()
	masterSmerd, err := testEnvironment.createSmerd(ctx, masterVersion)
	require.NoError(t, err)
	//TODO check master
	require.NotNil(t, masterSmerd)

	//testEnvironment.matreshkaApi
}

func Test_DeployGeneric(t *testing.T) {
	t.Run("postgres", func(t *testing.T) {
		t.Parallel()
		serviceName := getServiceName(t)

		timeout := uint32(5)
		command := "pg_isready -U postgres"

		createReq := &velez_api.CreateSmerd_Request{
			Name:      serviceName,
			ImageName: postgresImage,
			Env: map[string]string{
				"POSTGRES_HOST_AUTH_METHOD": "trust",
			},
			Labels: getExpectedLabels(),
			Healthcheck: &velez_api.Container_Healthcheck{
				Command:        &command,
				IntervalSecond: 2,
				TimeoutSecond:  &timeout,
				Retries:        3,
			},
			IgnoreConfig:  true,
			UseImagePorts: true,
		}

		ctx := context.Background()

		deployedSmerd, err := testEnvironment.createSmerd(ctx, createReq)
		require.NoError(t, err)

		expectedSmerd := &velez_api.Smerd{
			Name:      "/" + serviceName,
			ImageName: postgresImage,
			Status:    velez_api.Smerd_running,
			Labels:    getExpectedLabels(),
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
						strings.Replace(t.Name(), "/", "_", 1),
					},
				},
			},
		}

		assertSmerds(t, expectedSmerd, deployedSmerd)
	})

	t.Run("loki", func(t *testing.T) {
		t.Parallel()
		serviceName := getServiceName(t)
		ctx := context.Background()

		storeLokiConfigReq := &matreshka_api.StoreConfig_Request{
			Format:     matreshka_api.Format_yaml,
			ConfigName: serviceName,
			Config:     lokiConfig,
		}
		_, err := testEnvironment.app.Custom.MatreshkaClient.StoreConfig(ctx, storeLokiConfigReq)
		require.NoError(t, err)

		// TODO add plain config to matreshka
		createReq := &velez_api.CreateSmerd_Request{
			Name:      serviceName,
			ImageName: "grafana/loki:main-bc418c4",
			Settings: &velez_api.Container_Settings{
				Network: []*velez_api.NetworkBind{
					{
						NetworkName: "redsockru",
						Aliases:     []string{"loki"},
					},
				},
			},
			Restart: &velez_api.RestartPolicy{
				Type: velez_api.RestartPolicyType_always,
			},
			Config: &velez_api.MatreshkaConfigSpec{
				SystemPath: toolbox.ToPtr("/etc/loki/local-config.yaml"),
			},
		}

		deployedSmerd, err := testEnvironment.createSmerd(ctx, createReq)
		require.NoError(t, err)

		expectedSmerd := &velez_api.Smerd{
			Name:      "/" + serviceName,
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
					Aliases:     []string{"Test_DeployStr8_loki", "loki"},
				},
			},
			Labels: map[string]string{
				labels.CreatedWithVelezLabel: "true",
				integrationTest:              "true",
			},
			Env: map[string]string{
				"PATH":          "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/busybox",
				"SSL_CERT_FILE": "/etc/ssl/certs/ca-certificates.crt",
				"VERV_NAME":     "Test_DeployStr8_loki",
			},
			Binds: nil,
		}

		assertSmerds(t, expectedSmerd, deployedSmerd)
	})
}

func assertSmerds(t *testing.T, expected, actual *velez_api.Smerd) {
	require.NotEmpty(t, actual.Uuid)
	actual.Uuid = ""

	require.NotEmpty(t, actual.CreatedAt)
	actual.CreatedAt = nil

	require.Len(t, actual.Ports, len(expected.Ports))

	for idx, port := range actual.Ports {
		require.NotNil(t, port.ExposedTo)
		require.GreaterOrEqual(t, *port.ExposedTo, minPortToExposeTo)

		expected.Ports[idx].ExposedTo = port.ExposedTo

	}

	if !proto.Equal(expected, actual) {
		require.Equal(t, expected, actual)
	}
}
