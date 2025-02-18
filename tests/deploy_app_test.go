package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/godverv/Velez/pkg/velez_api"
)

const (
	helloWorldAppImage = "godverv/hello_world:v0.0.14"
	postgresImage      = "postgres:16"
)

func Test_DeployVerv(t *testing.T) {
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

func Test_DeployStr8(t *testing.T) {
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

		require.NotNil(t, deployedSmerd)
		require.NotEmpty(t, deployedSmerd.Uuid)
		require.NotEmpty(t, deployedSmerd.CreatedAt)

		deployedSmerd.Uuid = ""
		deployedSmerd.CreatedAt = nil

		expectedSmerd := &velez_api.Smerd{
			Name:      "/" + serviceName,
			ImageName: postgresImage,
			Status:    velez_api.Smerd_running,
			Labels:    getExpectedLabels(),
			Ports: []*velez_api.Port{
				{
					ServicePortNumber: 0,
					Protocol:          0,
					ExposedTo:         nil,
				},
			},
		}

		require.Equal(t, 1, len(deployedSmerd.Ports))
		require.Equal(t, uint32(5432), deployedSmerd.Ports[0].ServicePortNumber)
		require.Equal(t, velez_api.Port_tcp, deployedSmerd.Ports[0].Protocol)
		require.NotNil(t, *deployedSmerd.Ports[0].ExposedTo)
		require.GreaterOrEqual(t, *deployedSmerd.Ports[0].ExposedTo, minPortToExposeTo)

		deployedSmerd.Ports = nil
		expectedSmerd.Ports = nil

		if !proto.Equal(expectedSmerd, deployedSmerd) {
			require.Equal(t, expectedSmerd, deployedSmerd)
		}
	})
}
