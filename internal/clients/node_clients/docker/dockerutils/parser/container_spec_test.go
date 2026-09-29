package parser

import (
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func Test_ToCreateRequest_MapsInspectedContainer(t *testing.T) {
	t.Parallel()

	info := container.InspectResponse{
		Config: &container.Config{
			Image: "nginx:1.27",
			Env:   []string{"FOO=bar"},
			Healthcheck: &container.HealthConfig{
				Test:     []string{"CMD-SHELL", "curl -f localhost"},
				Interval: 5 * time.Second,
				Timeout:  2 * time.Second,
				Retries:  3,
			},
		},
		ContainerJSONBase: &container.ContainerJSONBase{
			HostConfig: &container.HostConfig{
				RestartPolicy: container.RestartPolicy{Name: container.RestartPolicyAlways},
			},
		},
	}

	req := ToCreateRequest("web", info)

	require.Equal(t, "web", req.GetName())
	require.Equal(t, "nginx:1.27", req.GetImageName())
	require.Equal(t, map[string]string{"FOO": "bar"}, req.GetEnv())
	require.Equal(t, "curl -f localhost", req.GetHealthcheck().GetCommand())
	require.Equal(t, uint32(5), req.GetHealthcheck().GetIntervalSecond())
	require.Equal(t, velez_api.RestartPolicyType_always, req.GetRestart().GetType())
	require.Nil(t, req.GetLabels())
}

func Test_ToHealthcheck_NoTestMeansNoHealthcheck(t *testing.T) {
	t.Parallel()

	require.Nil(t, ToHealthcheck(nil))
	require.Nil(t, ToHealthcheck(&container.HealthConfig{}))
}

func Test_ToRestartPolicy_Cases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   container.RestartPolicy
		want velez_api.RestartPolicyType
	}{
		{"empty is no", container.RestartPolicy{}, velez_api.RestartPolicyType_no},
		{
			"on failure",
			container.RestartPolicy{Name: container.RestartPolicyOnFailure},
			velez_api.RestartPolicyType_on_failure,
		},
		{"always", container.RestartPolicy{Name: container.RestartPolicyAlways}, velez_api.RestartPolicyType_always},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, ToRestartPolicy(tc.in).GetType())
		})
	}
}
