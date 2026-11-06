package selfswap

import (
	"os"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	testTargetId   = "abc123"
	testVelezName  = "velez"
	testVelezImage = "velez:old"
)

func Test_settingsFromEnv(t *testing.T) {
	cases := []struct {
		name    string
		env     map[string]string
		want    settings
		wantErr bool
	}{
		{
			name: "defaults when only target id is set",
			env: map[string]string{
				EnvTargetContainerId: testTargetId,
			},
			want: settings{
				targetContainerId: testTargetId,
				healthTimeout:     DefaultHealthTimeout,
				rollbackWindow:    DefaultRollbackWindow,
			},
		},
		{
			name: "all values overridden",
			env: map[string]string{
				EnvTargetContainerId: testTargetId,
				EnvImage:             "velez:v2",
				EnvHealthTimeout:     "5",
				EnvRollbackWindow:    "30",
			},
			want: settings{
				targetContainerId: testTargetId,
				image:             "velez:v2",
				healthTimeout:     time.Second * 5,
				rollbackWindow:    time.Second * 30,
			},
		},
		{
			name:    "missing target id is an error",
			env:     map[string]string{},
			wantErr: true,
		},
		{
			name: "unparsable health timeout is an error",
			env: map[string]string{
				EnvTargetContainerId: testTargetId,
				EnvHealthTimeout:     "not-a-number",
			},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(EnvTargetContainerId, "")
			t.Setenv(EnvImage, "")
			t.Setenv(EnvHealthTimeout, "")
			t.Setenv(EnvRollbackWindow, "")

			for k, v := range tc.env {
				t.Setenv(k, v)
			}

			got, err := settingsFromEnv()
			if tc.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func fixtureInspect() container.InspectResponse {
	return container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{
			Name:       "/velez",
			HostConfig: &container.HostConfig{},
		},
		Config: &container.Config{
			Image: testVelezImage,
			Env: []string{
				"PATH=/usr/bin",
				EnvSelfSwap + "=1",
				EnvTargetContainerId + "=deadbeef",
				"UNRELATED=true",
			},
			Labels: map[string]string{
				labels.SwapperLabel: "true",
				"verv.service":      testVelezName,
			},
		},
		NetworkSettings: &container.NetworkSettings{
			Networks: map[string]*network.EndpointSettings{
				"bridge": {
					Aliases: []string{testVelezName},
					Links:   []string{"db:db"},
				},
			},
		},
	}
}

func Test_cloneConfig(t *testing.T) {
	t.Parallel()

	inspect := fixtureInspect()

	cfg, hostCfg, netCfg := cloneConfig(&inspect, "velez:new")

	require.Equal(t, "velez:new", cfg.Image)
	require.Equal(t, []string{"PATH=/usr/bin", "UNRELATED=true"}, cfg.Env)
	require.Equal(t, map[string]string{"verv.service": testVelezName}, cfg.Labels)

	// The clone must not alias the original maps: mutating one side
	// (Docker create may normalise values in place) must not touch the
	// inspect result a rollback later relies on.
	cfg.Env[0] = "MUTATED"
	require.Equal(t, "PATH=/usr/bin", inspect.Config.Env[0])

	require.NotNil(t, hostCfg)
	require.NotNil(t, netCfg)
	require.Contains(t, netCfg.EndpointsConfig, "bridge")
	require.Equal(t, []string{testVelezName}, netCfg.EndpointsConfig["bridge"].Aliases)
}

func Test_containerName(t *testing.T) {
	t.Parallel()

	require.Equal(t, testVelezName, containerName("/"+testVelezName))
	require.Equal(t, testVelezName, containerName(testVelezName))
}

func Test_hostApiPort(t *testing.T) {
	t.Parallel()

	bindings := nat.PortMap{
		nat.Port(apiContainerPort): []nat.PortBinding{
			{HostPort: "54990"},
		},
	}

	hostConfig := &container.HostConfig{PortBindings: bindings}
	require.Equal(t, "54990", hostApiPort(hostConfig))

	emptyHostConfig := &container.HostConfig{}
	require.Equal(t, "53890", hostApiPort(emptyHostConfig))
}

func discardLogger() zerolog.Logger {
	return zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).
		With().Timestamp().Str("component", "selfswap-test").Logger()
}
